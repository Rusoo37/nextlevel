package services

import (
	"nextlevel/models"
	"time"
)

// GenerarHorariosLibres devuelve los strings de horarios (ej: "10:00") disponibles para una fecha,
// filtrando tanto los turnos normales ocupados como los turnos fijos de Ramón.
func GenerarHorariosLibres(fecha time.Time, turnosOcupados []models.Turno, turnosFijos []models.TurnoFijo) []string {
	var disponibles []string

	// Sacamos qué día de la semana es (0=Domingo, 1=Lunes, 2=Martes... 6=Sábado)
	diaSemana := int(fecha.Weekday())

	// 1. Regla de negocio: Ramón trabaja solo de Martes (2) a Sábado (6)
	if diaSemana == int(time.Sunday) || diaSemana == int(time.Monday) {
		return disponibles // Devuelve lista vacía, no hay turnos disponibles
	}

	// 2. Definir horario de atención de ese día arrando a las 9:00 y cerrando a las 19:00
	horaApertura := time.Date(fecha.Year(), fecha.Month(), fecha.Day(), 9, 0, 0, 0, fecha.Location())
	horaCierre := time.Date(fecha.Year(), fecha.Month(), fecha.Day(), 19, 0, 0, 0, fecha.Location())

	// Mapa para buscar y bloquear horarios rápidamente (clave: "HH:MM")
	horariosBloqueados := make(map[string]bool)

	// 3. Bloqueamos los turnos normales que ya están ocupados (y pagados/reservados) en esa fecha
	for _, t := range turnosOcupados {
		horaTexto := t.FechaHoraInicio.Format("15:04") // Formato 24hs (ej: "14:00")
		horariosBloqueados[horaTexto] = true
	}

	// 4. Bloqueamos los turnos fijos que caigan en este día de la semana
	for _, tf := range turnosFijos {
		if tf.DiaSemana == diaSemana && tf.Activo {
			horariosBloqueados[tf.Hora] = true
		}
	}

	ahora := time.Now()
	esHoy := fecha.Year() == ahora.Year() && fecha.Month() == ahora.Month() && fecha.Day() == ahora.Day()

	bloqueActual := horaApertura
	for bloqueActual.Before(horaCierre) {
		horaBloqueTexto := bloqueActual.Format("15:04")

		if esHoy && bloqueActual.Before(ahora) {
			// Si es hoy y el bloque ya pasó, NO lo mostramos
			bloqueActual = bloqueActual.Add(1 * time.Hour)
			continue
		}

		// Si el bloque está libre, lo agregamos
		if !horariosBloqueados[horaBloqueTexto] {
			disponibles = append(disponibles, horaBloqueTexto)
		}

		bloqueActual = bloqueActual.Add(1 * time.Hour)
	}

	return disponibles
}
