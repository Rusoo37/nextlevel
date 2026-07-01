package repository

import (
	"database/sql"
)

// InicializarTablas se encarga del DDL (Data Definition Language) de la app
func InicializarTablas(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS turnos (
		id SERIAL PRIMARY KEY,
		fecha_hora_inicio TIMESTAMP NOT NULL,
		fecha_hora_fin TIMESTAMP NOT NULL,
		nombre_cliente VARCHAR(100),
		telefono VARCHAR(20),
		email VARCHAR(100),
		estado VARCHAR(30) NOT NULL,
		monto_senado DECIMAL(10,2) DEFAULT 0,
		id_mercadopago VARCHAR(100),
		creado_en TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS configuracion (
		id SERIAL PRIMARY KEY,
		precio_turno DECIMAL(10,2) NOT NULL,
		porcentaje_sena INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS excepciones_calendario (
		id SERIAL PRIMARY KEY,
		fecha DATE NOT NULL,
		motivo VARCHAR(200)
	);

	CREATE TABLE IF NOT EXISTS turnos_fijos (
		id SERIAL PRIMARY KEY,
		dia_semana INTEGER NOT NULL, -- 2=Martes, 3=Miércoles... 6=Sábado (Estándar Go)
		hora VARCHAR(5) NOT NULL,    -- Ej: "17:00"
		nombre_cliente VARCHAR(100),
		activo BOOLEAN DEFAULT TRUE
	);
	
	-- Índice único para evitar concurrencia en turnos:
	CREATE UNIQUE INDEX IF NOT EXISTS idx_turno_unico 
	ON turnos (fecha_hora_inicio) 
	WHERE estado IN ('CONFIRMADO', 'PENDIENTE_PAGO', 'MANUAL');
	`

	_, err := db.Exec(query)
	return err
}
