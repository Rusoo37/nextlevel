package repository

import (
	"database/sql"
	"nextlevel/models"
	"time"
)

// ObtenerTurnosPorFecha busca turnos ocupados para un día específico
func ObtenerTurnosPorFecha(db *sql.DB, fecha time.Time) ([]models.Turno, error) {
	// Definimos el inicio y fin del día
	inicioDia := time.Date(fecha.Year(), fecha.Month(), fecha.Day(), 0, 0, 0, 0, fecha.Location())
	finDia := inicioDia.Add(24 * time.Hour)

	query := `
		SELECT id, fecha_hora_inicio, fecha_hora_fin, estado 
		FROM turnos 
		WHERE fecha_hora_inicio >= $1 AND fecha_hora_inicio < $2 
		AND estado IN ('CONFIRMADO', 'PENDIENTE_PAGO', 'MANUAL')
	`

	rows, err := db.Query(query, inicioDia, finDia)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var turnos []models.Turno
	for rows.Next() {
		var t models.Turno
		if err := rows.Scan(&t.ID, &t.FechaHoraInicio, &t.FechaHoraFin, &t.Estado); err != nil {
			return nil, err
		}
		turnos = append(turnos, t)
	}
	return turnos, nil
}

// ObtenerTurnosFijosPorDia trae los turnos recurrentes activos para un día de la semana (ej: 5 = Viernes)
func ObtenerTurnosFijosPorDia(db *sql.DB, diaSemana int) ([]models.TurnoFijo, error) {
	query := `
		SELECT id, dia_semana, hora, nombre_cliente, activo 
		FROM turnos_fijos 
		WHERE dia_semana = $1 AND activo = TRUE
	`

	rows, err := db.Query(query, diaSemana)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fijos []models.TurnoFijo
	for rows.Next() {
		var tf models.TurnoFijo
		if err := rows.Scan(&tf.ID, &tf.DiaSemana, &tf.Hora, &tf.NombreCliente, &tf.Activo); err != nil {
			return nil, err
		}
		fijos = append(fijos, tf)
	}
	return fijos, nil
}
