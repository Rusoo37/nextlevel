package repository

import (
	"database/sql"
	"time"
)

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
