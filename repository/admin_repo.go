package repository

import (
	"database/sql"
	"sort"
	"time"
)

type TurnoAdmin struct {
	ID            int    `json:"id"`
	Hora          string `json:"hora"`
	NombreCliente string `json:"nombre_cliente"`
	Telefono      string `json:"telefono"`
	Estado        string `json:"estado"`
}

type TurnoFijo struct {
	ID            int    `json:"id"`
	DiaSemana     int    `json:"dia_semana"`
	Hora          string `json:"hora"`
	NombreCliente string `json:"nombre_cliente"`
	Activo        bool   `json:"activo"`
}

type Configuracion struct {
	PrecioTurno    float64 `json:"precio_turno"`
	PorcentajeSena int     `json:"porcentaje_sena"`
}

func ObtenerTurnosDelDia(db *sql.DB, fecha string) ([]TurnoAdmin, error) {
	// 1. Buscamos los turnos normales (reservas de clientes)
	queryNormales := `
		SELECT id, fecha_hora_inicio, nombre_cliente, telefono, estado 
		FROM turnos 
		WHERE DATE(fecha_hora_inicio) = $1
		AND estado != 'CANCELADO'
		ORDER BY fecha_hora_inicio ASC
	`
	rows, err := db.Query(queryNormales, fecha)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Usamos un mapa para saber qué horas ya están ocupadas por un turno normal
	turnosMap := make(map[string]TurnoAdmin)
	var turnos []TurnoAdmin

	for rows.Next() {
		var t TurnoAdmin
		var fechaHora time.Time
		err := rows.Scan(&t.ID, &fechaHora, &t.NombreCliente, &t.Telefono, &t.Estado)
		if err != nil {
			continue
		}
		t.Hora = fechaHora.Format("15:04")
		turnosMap[t.Hora] = t
		turnos = append(turnos, t)
	}

	// 2. Descubrimos qué día de la semana es la fecha que pidió Ramón
	fechaParseada, _ := time.Parse("2006-01-02", fecha)
	diaSemana := int(fechaParseada.Weekday()) // 0=Dom, 1=Lun, ..., 4=Jueves, etc.

	// 3. Buscamos los turnos fijos que caen en ese día de la semana
	queryFijos := `
		SELECT hora, nombre_cliente 
		FROM turnos_fijos 
		WHERE dia_semana = $1 AND activo = TRUE
	`
	rowsFijos, err := db.Query(queryFijos, diaSemana)
	if err == nil {
		defer rowsFijos.Close()
		for rowsFijos.Next() {
			var tfHora string
			var nombreNull sql.NullString

			rowsFijos.Scan(&tfHora, &nombreNull)

			// Le ponemos un nombre por defecto si está vacío
			tfNombre := "Bloqueo Fijo"
			if nombreNull.Valid && nombreNull.String != "" {
				tfNombre = nombreNull.String
			}

			// Solo lo agregamos si no hay un turno normal pisando esa misma hora
			if _, existe := turnosMap[tfHora]; !existe {
				tFijo := TurnoAdmin{
					ID:            0,
					Hora:          tfHora,
					NombreCliente: tfNombre,
					Telefono:      "-",
					Estado:        "FIJO",
				}
				turnos = append(turnos, tFijo)
			}
		}
	}

	// 4. Como agregamos los fijos al final de la lista, volvemos a ordenar todo por hora
	sort.Slice(turnos, func(i, j int) bool {
		return turnos[i].Hora < turnos[j].Hora
	})

	return turnos, nil
}

func BloquearHorarioManual(db *sql.DB, fechaHora time.Time) error {
	fin := fechaHora.Add(30 * time.Minute)
	query := `
		INSERT INTO turnos (fecha_hora_inicio, fecha_hora_fin, nombre_cliente, telefono, estado)
		VALUES ($1, $2, 'BLOQUEADO', '-', 'MANUAL')
	`
	_, err := db.Exec(query, fechaHora, fin)
	return err
}

// ObtenerTodosLosTurnosFijos trae los bloqueos activos
func ObtenerTodosLosTurnosFijos(db *sql.DB) ([]TurnoFijo, error) {
	query := `
		SELECT id, dia_semana, hora, nombre_cliente, activo 
		FROM turnos_fijos 
		WHERE activo = TRUE 
		ORDER BY dia_semana, hora
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fijos []TurnoFijo
	for rows.Next() {
		var tf TurnoFijo
		// Manejamos el NULL por si el nombre está vacío en la BD
		var nombre sql.NullString

		err := rows.Scan(&tf.ID, &tf.DiaSemana, &tf.Hora, &nombre, &tf.Activo)
		if err != nil {
			continue
		}

		if nombre.Valid {
			tf.NombreCliente = nombre.String
		} else {
			tf.NombreCliente = "Bloqueado"
		}

		fijos = append(fijos, tf)
	}
	return fijos, nil
}

// GuardarTurnoFijo inserta una nueva regla respetando tus campos
func GuardarTurnoFijo(db *sql.DB, dia int, hora, nombre string) error {
	query := `
		INSERT INTO turnos_fijos (dia_semana, hora, nombre_cliente, activo) 
		VALUES ($1, $2, $3, TRUE)
	`
	_, err := db.Exec(query, dia, hora, nombre)
	return err
}

// EliminarTurnoFijo borra una regla
func EliminarTurnoFijo(db *sql.DB, dia int, hora string) error {
	query := `DELETE FROM turnos_fijos WHERE dia_semana = $1 AND hora = $2`
	_, err := db.Exec(query, dia, hora)
	return err
}

// para los precios de los turnos

// ObtenerConfiguracion lee los precios actuales
func ObtenerConfiguracion(db *sql.DB) (Configuracion, error) {
	var config Configuracion
	// Buscamos la primera fila de configuración
	query := `SELECT precio_turno, porcentaje_sena FROM configuracion LIMIT 1`
	err := db.QueryRow(query).Scan(&config.PrecioTurno, &config.PorcentajeSena)

	if err != nil {
		if err == sql.ErrNoRows {
			// Si la tabla está vacía, devolvemos valores por defecto
			return Configuracion{PrecioTurno: 15000.00, PorcentajeSena: 50}, nil
		}
		return config, err
	}
	return config, nil
}

// ActualizarConfiguracion guarda los nuevos valores de Ramón
func ActualizarConfiguracion(db *sql.DB, precio float64, sena int) error {
	var cantidad int
	db.QueryRow(`SELECT COUNT(*) FROM configuracion`).Scan(&cantidad)

	// Si no hay nada, insertamos. Si ya existe, actualizamos.
	if cantidad == 0 {
		query := `INSERT INTO configuracion (precio_turno, porcentaje_sena) VALUES ($1, $2)`
		_, err := db.Exec(query, precio, sena)
		return err
	}

	query := `UPDATE configuracion SET precio_turno = $1, porcentaje_sena = $2`
	_, err := db.Exec(query, precio, sena)
	return err
}

// aca vamos a poner las excepciones como las vacaciones o algun dia que no quiera ir a laburar

type ExcepcionCalendario struct {
	ID     int    `json:"id"`
	Fecha  string `json:"fecha"`
	Motivo string `json:"motivo"`
}

// ObtenerExcepciones trae la lista de días bloqueados futuros
func ObtenerExcepciones(db *sql.DB) ([]ExcepcionCalendario, error) {
	// Solo traemos de hoy en adelante, lo pasado ya no importa
	query := `SELECT id, fecha, motivo FROM excepciones_calendario WHERE fecha >= CURRENT_DATE ORDER BY fecha ASC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var excepciones []ExcepcionCalendario
	for rows.Next() {
		var exc ExcepcionCalendario
		var fecha time.Time
		err := rows.Scan(&exc.ID, &fecha, &exc.Motivo)
		if err != nil {
			continue
		}
		exc.Fecha = fecha.Format("2006-01-02")
		excepciones = append(excepciones, exc)
	}
	return excepciones, nil
}

// GuardarExcepcion bloquea un día completo
func GuardarExcepcion(db *sql.DB, fecha, motivo string) error {
	query := `INSERT INTO excepciones_calendario (fecha, motivo) VALUES ($1, $2)`
	_, err := db.Exec(query, fecha, motivo)
	return err
}

// EliminarExcepcion borra un bloqueo de día completo
func EliminarExcepcion(db *sql.DB, id int) error {
	query := `DELETE FROM excepciones_calendario WHERE id = $1`
	_, err := db.Exec(query, id)
	return err
}

// EsDiaExcepcion devuelve true si el día está bloqueado (para que el cliente no pueda reservar)
func EsDiaExcepcion(db *sql.DB, fecha string) bool {
	var existe bool
	query := `SELECT EXISTS(SELECT 1 FROM excepciones_calendario WHERE fecha = $1)`
	err := db.QueryRow(query, fecha).Scan(&existe)
	if err != nil {
		return false
	}
	return existe
}

// CancelarTurnoBD cambia el estado de un turno específico a 'CANCELADO'
func CancelarTurnoBD(db *sql.DB, idTurno int) error {
	query := `UPDATE turnos SET estado = 'CANCELADO' WHERE id = $1`
	_, err := db.Exec(query, idTurno)
	return err
}
