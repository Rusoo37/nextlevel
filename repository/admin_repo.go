package repository

import (
	"database/sql"
	"time"
)

type TurnoAdmin struct {
	ID            int    `json:"id"`
	Hora          string `json:"hora"`
	NombreCliente string `json:"nombre_cliente"`
	Telefono      string `json:"telefono"`
	Estado        string `json:"estado"`
}

func ObtenerTurnosDelDia(db *sql.DB, fecha string) ([]TurnoAdmin, error) {
	// Traemos todos los turnos que existan para el día
	query := `
		SELECT id, fecha_hora_inicio, nombre_cliente, telefono, estado 
		FROM turnos 
		WHERE DATE(fecha_hora_inicio) = $1
		ORDER BY fecha_hora_inicio ASC
	`
	rows, err := db.Query(query, fecha)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var turnos []TurnoAdmin
	for rows.Next() {
		var t TurnoAdmin
		var fechaHora time.Time
		err := rows.Scan(&t.ID, &fechaHora, &t.NombreCliente, &t.Telefono, &t.Estado)
		if err != nil {
			continue
		}
		t.Hora = fechaHora.Format("15:04")
		turnos = append(turnos, t)
	}
	return turnos, nil
}

func BloquearHorarioManual(db *sql.DB, fechaHora time.Time) error {
	fin := fechaHora.Add(30 * time.Minute)
	query := `
		INSERT INTO turnos (fecha_hora_inicio, fecha_hora_fin, nombre_cliente, telefono, estado)
		VALUES ($1, $2, 'BLOQUEADO (Admin)', '-', 'MANUAL')
	`
	_, err := db.Exec(query, fechaHora, fin)
	return err
}
