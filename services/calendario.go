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

	// 1. Regla de negocio: Ramón no trabaja los Domingos
	if diaSemana == int(time.Sunday) {
		return disponibles // Devuelve lista vacía, no hay turnos disponibles
	}

	// 2. Definir horario de atención dinámico según el día
	var horaInicio, horaFin int

	switch diaSemana {
	case int(time.Monday):
		horaInicio = 12
		horaFin = 18 // El bucle frena antes de las 18, el último turno será 17:00
	case int(time.Saturday):
		horaInicio = 10
		horaFin = 14 // El bucle frena antes de las 14, el último turno será 13:00
	default:
		// Martes, Miércoles, Jueves y Viernes
		horaInicio = 10
		horaFin = 18 // El bucle frena antes de las 18, el último turno será 17:00
	}

	horaApertura := time.Date(fecha.Year(), fecha.Month(), fecha.Day(), horaInicio, 0, 0, 0, fecha.Location())
	horaCierre := time.Date(fecha.Year(), fecha.Month(), fecha.Day(), horaFin, 0, 0, 0, fecha.Location())

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

	loc, _ := time.LoadLocation("America/Argentina/Buenos_Aires")
	ahora := time.Now().In(loc)
	esHoy := fecha.Year() == ahora.Year() && fecha.Month() == ahora.Month() && fecha.Day() == ahora.Day()

	// 5. Generar los turnos bloque a bloque (saltos de 1 hora)
	bloqueActual := horaApertura
	for bloqueActual.Before(horaCierre) {
		horaBloqueTexto := bloqueActual.Format("15:04")

		bloqueActual = bloqueActual.In(loc)

		if esHoy && bloqueActual.Before(ahora) {
			// Si es hoy y el bloque ya pasó, saltamos
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
