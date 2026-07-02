package repository

import (
	"database/sql"
	"time"
)

// CrearReservaTemporal intenta guardar un turno bloqueándolo para que el cliente pague la seña. Devuelve el ID del turno insertado o un error.
func CrearReservaTemporal(db *sql.DB, inicio time.Time, nombre, telefono string, montoSena float64) (int, error) {
	fin := inicio.Add(30 * time.Minute)

	query := `
        INSERT INTO turnos (fecha_hora_inicio, fecha_hora_fin, nombre_cliente, telefono, estado, monto_senado)
        VALUES ($1, $2, $3, $4, 'PENDIENTE_PAGO', $5) 
        RETURNING id
    `

	var idInsertado int
	err := db.QueryRow(query, inicio, fin, nombre, telefono, montoSena).Scan(&idInsertado)
	if err != nil {
		return 0, err
	}

	return idInsertado, nil
}

// ConfirmarTurno cambia el estado a CONFIRMADO cuando el pago es exitoso
func ConfirmarTurno(db *sql.DB, idTurno int, idMercadoPago string) error {
	query := `
		UPDATE turnos 
		SET estado = 'CONFIRMADO', id_mercadopago = $1 
		WHERE id = $2 AND estado = 'PENDIENTE_PAGO'
	`
	_, err := db.Exec(query, idMercadoPago, idTurno)
	return err
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

// ObtenerDetalleTurno busca los datos de un turno por su ID para el recibo
func ObtenerDetalleTurno(db *sql.DB, id int) (string, string, error) {
	var nombre string
	var fecha time.Time

	query := `SELECT nombre_cliente, fecha_hora_inicio FROM turnos WHERE id = $1`
	err := db.QueryRow(query, id).Scan(&nombre, &fecha)

	if err != nil {
		return "", "", err
	}

	// Formateamos la fecha a formato argentino (Día/Mes/Año Hora:Minutos)
	fechaFormateada := fecha.Format("02/01/2006 15:04")
	return nombre, fechaFormateada, nil
}
