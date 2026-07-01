package services

import (
	"time"

	"nextlevel/models"
)

// GenerarHorariosLibres devuelve los strings de horarios (ej: "09:30") disponibles para una fecha,
// filtrando tanto los turnos normales ocupados como los turnos fijos de Ramón.
func GenerarHorariosLibres(fecha time.Time, turnosOcupados []models.Turno, turnosFijos []models.TurnoFijo) []string {
	var disponibles []string

	// Sacamos qué día de la semana es (0=Domingo, 1=Lunes, 2=Martes... 6=Sábado)
	diaSemana := int(fecha.Weekday())

	// 1. Regla de negocio: Ramón trabaja solo de Martes (2) a Sábado (6)
	if diaSemana == int(time.Sunday) || diaSemana == int(time.Monday) {
		return disponibles // Devuelve lista vacía, no hay turnos disponibles
	}

	// 2. Definir horario de atención de ese día (09:00 a 18:00)
	horaApertura := time.Date(fecha.Year(), fecha.Month(), fecha.Day(), 9, 0, 0, 0, fecha.Location())
	horaCierre := time.Date(fecha.Year(), fecha.Month(), fecha.Day(), 18, 0, 0, 0, fecha.Location())

	// Mapa para buscar y bloquear horarios rápidamente (clave: "HH:MM")
	horariosBloqueados := make(map[string]bool)

	// 3. Bloqueamos los turnos normales que ya están ocupados (y pagados/reservados) en esa fecha
	for _, t := range turnosOcupados {
		horaTexto := t.FechaHoraInicio.Format("15:04") // Formato 24hs (ej: "14:30")
		horariosBloqueados[horaTexto] = true
	}

	// 4. Bloqueamos los turnos fijos que caigan en este día de la semana
	for _, tf := range turnosFijos {
		// Si el turno fijo es para este día y está activo, lo bloqueamos
		if tf.DiaSemana == diaSemana && tf.Activo {
			horariosBloqueados[tf.Hora] = true
		}
	}

	// 5. Generar bloques de 30 minutos y filtrar los que están libres
	bloqueActual := horaApertura
	for bloqueActual.Before(horaCierre) {
		horaBloqueTexto := bloqueActual.Format("15:04")

		// Si la hora actual NO está en el mapa de bloqueados, significa que está libre
		if !horariosBloqueados[horaBloqueTexto] {
			disponibles = append(disponibles, horaBloqueTexto)
		}

		// Sumamos 30 minutos para el siguiente ciclo del bucle
		bloqueActual = bloqueActual.Add(30 * time.Minute)
	}

	return disponibles
}
