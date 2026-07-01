package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
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

// Disponibilidad maneja la petición GET /api/disponibilidad?fecha=YYYY-MM-DD
func (api *APIHandler) Disponibilidad(w http.ResponseWriter, r *http.Request) {
	// 1. Leer la fecha que manda el cliente por la URL
	fechaStr := r.URL.Query().Get("fecha")
	if fechaStr == "" {
		http.Error(w, "Falta el parámetro 'fecha'", http.StatusBadRequest)
		return
	}

	// 2. Parsear la fecha
	fecha, err := time.Parse("2006-01-02", fechaStr)
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
		// Acá a futuro podrías validar con una expresión regular (Regex)
		// que solo sean números, pero por ahora con que no esté vacío alcanza.
		http.Error(w, "El teléfono de contacto es obligatorio", http.StatusBadRequest)
		return
	}

	// Parseamos la hora que eligió el cliente (ej: "2026-07-15T10:30:00Z")
	fechaInicio, err := time.Parse(time.RFC3339, solicitud.FechaHoraInicio)
	if err != nil {
		http.Error(w, "Formato de fecha inválido", http.StatusBadRequest)
		return
	}

	montoSena := 7500.00
	idTurno, err := repository.CrearReservaTemporal(api.DB, fechaInicio, solicitud.NombreCliente, solicitud.Telefono, solicitud.Email)

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
func (api *APIHandler) WebhookMercadoPago(w http.ResponseWriter, r *http.Request) {
	// Mercado Pago nos manda el ID del evento (el pago) en el query string
	log.Printf("📥 RECIBIENDO NOTIFICACIÓN: %v", r.URL.Query())
	idPago := r.URL.Query().Get("data.id")

	if idPago == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 1. Acá deberías llamar a la API de Mercado Pago para consultar el pago
	// Pero para simplificar, vamos a asumir que recibimos el "ExternalReference"
	// que configuramos al crear el pago (que es el ID de TU turno).

	// Si usaste la librería de MP, deberías hacer algo como:
	// paymentClient := payment.NewClient(os.Getenv("MP_ACCESS_TOKEN"))
	// paymentData, _ := paymentClient.Get(idPago)
	// idTurno := paymentData.ExternalReference

	// HASTA QUE IMPLEMENTES LA CONSULTA REAL A MP,
	// necesitamos que el Webhook reciba el turno.
	// ¿Tu base de datos qué ID tiene guardado en 'id_mercadopago'?
	// Quizás el Webhook no está encontrando el turno porque no le estás pasando el ID.

	// CORRECCIÓN RÁPIDA:
	// Asegurate de que cuando creás la preferencia en services/pagos.go
	// hayas puesto: ExternalReference: fmt.Sprintf("%d", idTurno)

	// Y en el Webhook, simplemente actualizá el turno donde el estado sea PENDIENTE_PAGO
	// y el pago sea el que corresponde.

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
