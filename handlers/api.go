package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
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

	// Intentamos guardar en la base de datos
	idTurno, err := repository.CrearReservaTemporal(api.DB, fechaInicio, solicitud.NombreCliente, solicitud.Telefono, solicitud.Email)

	if err != nil {
		// Si da error, casi seguro es por nuestro blindaje de concurrencia
		http.Error(w, "El turno ya no está disponible, por favor elegí otro.", http.StatusConflict)
		return
	}

	// ¡Éxito! El turno está bloqueado.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mensaje":  "Turno pre-reservado con éxito. Pendiente de pago.",
		"id_turno": idTurno,
	})
}
