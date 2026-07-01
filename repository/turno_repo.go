package repository

import (
	"database/sql"
	"time"

	"nextlevel/models"
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

// CrearReservaTemporal intenta guardar un turno bloqueándolo para que el cliente pague
func CrearReservaTemporal(db *sql.DB, inicio time.Time, nombre, telefono, email string) (int, error) {
	// Calculamos el fin del turno (30 min después)
	fin := inicio.Add(30 * time.Minute)

	// Usamos un valor fijo para la seña por ahora (la mitad de 15.000)
	sena := 7500.00

	var idInsertado int
	query := `
		INSERT INTO turnos (fecha_hora_inicio, fecha_hora_fin, nombre_cliente, telefono, email, estado, monto_senado)
		VALUES ($1, $2, $3, $4, $5, 'PENDIENTE_PAGO', $6) 
		RETURNING id
	`

	// Intentamos insertar. Si Juan y María llegan a la vez, el índice único que creamos
	// hará que a uno de los dos le salte un error acá.
	err := db.QueryRow(query, inicio, fin, nombre, telefono, email, sena).Scan(&idInsertado)
	if err != nil {
		return 0, err
	}

	return idInsertado, nil
}

// CancelarTurnosExpirados cambia a CANCELADO los turnos que no se pagaron a tiempo
func CancelarTurnosExpirados(db *sql.DB, minutos int) (int64, error) {
	// Calculamos qué hora era hace 'X' minutos atrás
	tiempoLimite := time.Now().Add(-time.Duration(minutos) * time.Minute)

	query := `
		UPDATE turnos 
		SET estado = 'CANCELADO' 
		WHERE estado = 'PENDIENTE_PAGO' 
		AND creado_en < $1
	`

	resultado, err := db.Exec(query, tiempoLimite)
	if err != nil {
		return 0, err
	}

	// Devolvemos cuántas filas se cancelaron para llevar un control
	return resultado.RowsAffected()
}
