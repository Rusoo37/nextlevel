package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"nextlevel/models"
	"nextlevel/repository"
	"nextlevel/services"
)

// APIHandler estructura que guarda la conexión a la DB para usarla en las rutas
type APIHandler struct {
	DB *sql.DB
}

var telefonoRegex = regexp.MustCompile(`^[0-9+() ]{7,20}$`)

// Disponibilidad maneja la petición GET /api/disponibilidad?fecha=YYYY-MM-DD
func (api *APIHandler) Disponibilidad(w http.ResponseWriter, r *http.Request) {

	// 1. Leer la fecha que manda el cliente por la URL
	fechaStr := r.URL.Query().Get("fecha")
	if fechaStr == "" {
		http.Error(w, "Falta el parámetro 'fecha'", http.StatusBadRequest)
		return
	}
	if repository.EsDiaExcepcion(api.DB, fechaStr) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"fecha":       fechaStr,
			"disponibles": []string{}, // Lista vacía, no hay nada libre
		})
		return
	}

	// 2. Parsear la fecha
	loc, _ := time.LoadLocation("America/Argentina/Buenos_Aires")
	fecha, err := time.ParseInLocation("2006-01-02", fechaStr, loc)

	if err != nil {
		http.Error(w, "Formato de fecha inválido. Usar YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	// 3. Buscar turnos ocupados en la DB
	turnosOcupados, err := repository.ObtenerTurnosPorFecha(api.DB, fecha)
	if err != nil {
		http.Error(w, "Error al consultar turnos ocupados", http.StatusInternalServerError)
		return
	}

	// 4. Buscar turnos fijos en la DB para ese día de la semana
	turnosFijos, err := repository.ObtenerTurnosFijosPorDia(api.DB, int(fecha.Weekday()))
	if err != nil {
		http.Error(w, "Error al consultar turnos fijos", http.StatusInternalServerError)
		return
	}

	// 5. Usar nuestro servicio para calcular los libres
	horariosLibres := services.GenerarHorariosLibres(fecha, turnosOcupados, turnosFijos)

	// 6. Devolver la respuesta en formato JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"fecha":       fechaStr,
		"disponibles": horariosLibres,
	})
}

// Reservar maneja la petición POST /api/reservar
func (api *APIHandler) Reservar(w http.ResponseWriter, r *http.Request) {
	// Solo aceptamos método POST
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Decodificamos el JSON que manda el cliente
	var solicitud models.SolicitudReserva
	if err := json.NewDecoder(r.Body).Decode(&solicitud); err != nil {
		http.Error(w, "Error al leer los datos", http.StatusBadRequest)
		return
	}

	// Limpiamos espacios en blanco al principio y al final
	solicitud.NombreCliente = strings.TrimSpace(solicitud.NombreCliente)
	solicitud.Telefono = strings.TrimSpace(solicitud.Telefono)

	// Verificamos que no estén vacíos
	if solicitud.NombreCliente == "" {
		http.Error(w, "El nombre del cliente es obligatorio", http.StatusBadRequest)
		return
	}

	if solicitud.Telefono == "" {
		http.Error(w, "El teléfono de contacto es obligatorio", http.StatusBadRequest)
		return
	}
	if !telefonoRegex.MatchString(solicitud.Telefono) {
		http.Error(w, "El número de teléfono no parece válido. Por favor ingresá solo números.", http.StatusBadRequest)
		return
	}

	// Parseamos la hora que eligió el cliente (ej: "2026-07-15T10:30:00Z")
	fechaInicio, err := time.Parse(time.RFC3339, solicitud.FechaHoraInicio)
	if err != nil {
		http.Error(w, "Formato de fecha inválido", http.StatusBadRequest)
		return
	}
	config, errConfig := repository.ObtenerConfiguracion(api.DB)
	if errConfig != nil {
		http.Error(w, "Error interno leyendo configuración de precios", http.StatusInternalServerError)
		return
	}

	montoSena := (config.PrecioTurno * float64(config.PorcentajeSena)) / 100.0
	idTurno, err := repository.CrearReservaTemporal(api.DB, fechaInicio, solicitud.NombreCliente, solicitud.Telefono, montoSena)

	if err != nil {
		http.Error(w, "El turno ya no está disponible, por favor elegí otro.", http.StatusConflict)
		return
	}

	// Generamos el link de pago en Mercado Pago
	linkPago, err := services.GenerarLinkDePago(idTurno, montoSena, solicitud.NombreCliente)
	if err != nil {
		// Imprimimos el error real en la consola del servidor para poder debuggear
		log.Printf("ERROR de Mercado Pago: %v\n", err)

		repository.CancelarTurnosExpirados(api.DB, 0)
		http.Error(w, "Error al generar el pago. Intente nuevamente.", http.StatusInternalServerError)
		return
	}

	// El turno está bloqueado y tenemos el link.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mensaje":   "Turno pre-reservado con éxito.",
		"id_turno":  idTurno,
		"link_pago": linkPago, // El frontend debe redirigir al usuario a esta URL
	})
}

// WebhookMercadoPago recibe las notificaciones de pagos de MP
// WebhookMercadoPago recibe las notificaciones de pagos de MP
func (api *APIHandler) WebhookMercadoPago(w http.ResponseWriter, r *http.Request) {
	// 1. Agarramos el ID del pago que nos manda Mercado Pago
	idPago := r.URL.Query().Get("data.id")
	if idPago == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	log.Printf("📥 RECIBIENDO NOTIFICACIÓN DEL PAGO ID: %s", idPago)

	// 2. Le preguntamos a Mercado Pago: "Che, ¿qué onda este pago?"
	url := "https://api.mercadopago.com/v1/payments/" + idPago
	req, _ := http.NewRequest("GET", url, nil)

	tokenMP := os.Getenv("MP_ACCESS_TOKEN")
	req.Header.Add("Authorization", "Bearer "+tokenMP)
	// Usamos tu mismo token de prueba para autorizar la pregunta

	clienteHTTP := &http.Client{}
	respuestaMP, err := clienteHTTP.Do(req)
	if err != nil {
		log.Printf("Error consultando a MP: %v", err)
		w.WriteHeader(http.StatusOK)
		return
	}
	defer respuestaMP.Body.Close()

	// 3. Leemos la respuesta de Mercado Pago
	var dataPago struct {
		ExternalReference string `json:"external_reference"`
		Status            string `json:"status"`
	}
	json.NewDecoder(respuestaMP.Body).Decode(&dataPago)

	// 4. Verificamos si el pago fue aprobado y sacamos nuestro ID de turno
	if dataPago.Status == "approved" && dataPago.ExternalReference != "" {
		idTurno, _ := strconv.Atoi(dataPago.ExternalReference)

		// 5. ¡ACÁ ESTÁ LA MAGIA! Le avisamos a tu base de datos
		err = repository.ConfirmarTurno(api.DB, idTurno, idPago)

		if err != nil {
			log.Printf("🚨 Error al confirmar turno en BD: %v", err)
		} else {
			log.Printf("✅ ¡ÉXITO! Turno %d confirmado en la base de datos.", idTurno)
		}
	} else {
		log.Printf("⚠️ El pago %s no está aprobado o no tiene referencia. Estado: %s", idPago, dataPago.Status)
	}

	// 6. Siempre hay que responderle 200 OK a MP rápido para que deje de avisar
	w.WriteHeader(http.StatusOK)
}

// DetalleTurno devuelve los datos de un turno específico en formato JSON
func (api *APIHandler) DetalleTurno(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	idTurno, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	nombre, fecha, err := repository.ObtenerDetalleTurno(api.DB, idTurno)
	if err != nil {
		http.Error(w, "Turno no encontrado", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"nombre": nombre,
		"fecha":  fecha,
	})
}

// AdminTurnosHandler devuelve la lista de turnos para el panel de control
func (api *APIHandler) AdminTurnosHandler(w http.ResponseWriter, r *http.Request) {
	fecha := r.URL.Query().Get("fecha")
	if fecha == "" {
		// Si no mandan fecha, por defecto buscamos los de "hoy"
		fecha = time.Now().Format("2006-01-02")
	}

	turnos, err := repository.ObtenerTurnosDelDia(api.DB, fecha)
	if err != nil {
		http.Error(w, "Error al obtener los turnos", http.StatusInternalServerError)
		return
	}

	// Si no hay turnos, devolvemos un arreglo vacío en lugar de "null"
	if turnos == nil {
		turnos = []repository.TurnoAdmin{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(turnos)
}

// BloquearHorarioAdmin maneja la petición POST para bloquear un turno a mano
func (api *APIHandler) BloquearHorarioAdmin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Creamos una estructura rápida solo para leer la hora que manda el admin
	var peticion struct {
		FechaHora string `json:"fecha_hora"`
	}

	if err := json.NewDecoder(r.Body).Decode(&peticion); err != nil {
		http.Error(w, "Error al leer los datos", http.StatusBadRequest)
		return
	}

	// Parseamos la fecha (ej: "2026-07-16T15:00:00Z")
	fechaParseada, err := time.Parse(time.RFC3339, peticion.FechaHora)
	if err != nil {
		http.Error(w, "Formato de fecha inválido", http.StatusBadRequest)
		return
	}

	// Intentamos bloquear el horario
	err = repository.BloquearHorarioManual(api.DB, fechaParseada)
	if err != nil {
		http.Error(w, "Ese horario ya está ocupado o bloqueado.", http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"mensaje": "Horario bloqueado exitosamente",
	})
}

// ConfigFijosHandler maneja la lectura, creación y eliminación de turnos fijos
func (api *APIHandler) ConfigFijosHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodGet {
		fijos, err := repository.ObtenerTodosLosTurnosFijos(api.DB)
		if err != nil {
			http.Error(w, "Error al obtener configuración", http.StatusInternalServerError)
			return
		}
		if fijos == nil {
			fijos = []repository.TurnoFijo{}
		}
		json.NewEncoder(w).Encode(fijos)
		return
	}

	if r.Method == http.MethodPost {
		var peticion struct {
			Accion        string `json:"accion"` // "guardar" o "eliminar"
			DiaSemana     int    `json:"dia_semana"`
			Hora          string `json:"hora"`
			NombreCliente string `json:"nombre_cliente"`
		}

		if err := json.NewDecoder(r.Body).Decode(&peticion); err != nil {
			http.Error(w, "Datos inválidos", http.StatusBadRequest)
			return
		}

		if peticion.DiaSemana < 0 || peticion.DiaSemana > 6 {
			http.Error(w, "Día de la semana inválido", http.StatusBadRequest)
			return
		}

		var err error
		if peticion.Accion == "guardar" {
			// Si Ramón no puso nombre, le ponemos uno por defecto
			if peticion.NombreCliente == "" {
				peticion.NombreCliente = "Fijo/Personal"
			}
			err = repository.GuardarTurnoFijo(api.DB, peticion.DiaSemana, peticion.Hora, peticion.NombreCliente)
		} else if peticion.Accion == "eliminar" {
			err = repository.EliminarTurnoFijo(api.DB, peticion.DiaSemana, peticion.Hora)
		} else {
			http.Error(w, "Acción desconocida", http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, "Error al procesar la regla", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"mensaje": "Configuración actualizada"})
		return
	}

	http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
}

// ConfigPreciosHandler maneja la vista y actualización de precios
func (api *APIHandler) ConfigPreciosHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Si es GET, devolvemos la configuración actual
	if r.Method == http.MethodGet {
		config, err := repository.ObtenerConfiguracion(api.DB)
		if err != nil {
			http.Error(w, "Error al obtener configuración", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(config)
		return
	}

	// Si es POST, Ramón está guardando nuevos precios
	if r.Method == http.MethodPost {
		var peticion struct {
			Precio float64 `json:"precio_turno"`
			Sena   int     `json:"porcentaje_sena"`
		}

		if err := json.NewDecoder(r.Body).Decode(&peticion); err != nil {
			http.Error(w, "Datos inválidos", http.StatusBadRequest)
			return
		}

		// Validaciones básicas de negocio
		if peticion.Precio <= 0 || peticion.Sena < 0 || peticion.Sena > 100 {
			http.Error(w, "Valores fuera de rango", http.StatusBadRequest)
			return
		}

		err := repository.ActualizarConfiguracion(api.DB, peticion.Precio, peticion.Sena)
		if err != nil {
			http.Error(w, "Error al guardar los precios", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"mensaje": "Precios actualizados"})
		return
	}

	http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
}

// ConfigExcepcionesHandler maneja la vista y actualización de vacaciones/feriados
func (api *APIHandler) ConfigExcepcionesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodGet {
		excepciones, err := repository.ObtenerExcepciones(api.DB)
		if err != nil {
			http.Error(w, "Error al obtener excepciones", http.StatusInternalServerError)
			return
		}
		if excepciones == nil {
			excepciones = []repository.ExcepcionCalendario{}
		}
		json.NewEncoder(w).Encode(excepciones)
		return
	}

	if r.Method == http.MethodPost {
		var peticion struct {
			Accion string `json:"accion"` // "guardar" o "eliminar"
			ID     int    `json:"id"`
			Fecha  string `json:"fecha"`
			Motivo string `json:"motivo"`
		}

		if err := json.NewDecoder(r.Body).Decode(&peticion); err != nil {
			http.Error(w, "Datos inválidos", http.StatusBadRequest)
			return
		}

		var err error
		if peticion.Accion == "guardar" {
			if peticion.Motivo == "" {
				peticion.Motivo = "Cerrado / Vacaciones"
			}
			err = repository.GuardarExcepcion(api.DB, peticion.Fecha, peticion.Motivo)
		} else if peticion.Accion == "eliminar" {
			err = repository.EliminarExcepcion(api.DB, peticion.ID)
		} else {
			http.Error(w, "Acción desconocida", http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, "Error al procesar la excepción", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"mensaje": "Calendario actualizado"})
		return
	}

	http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
}
