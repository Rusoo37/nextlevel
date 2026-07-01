package config

import (
	"database/sql"

	_ "github.com/lib/pq"
)

const dbURL = "postgres://ramon_admin:supersecretpassword@localhost:5432/nextlevel_db?sslmode=disable"

// ObtenerConexion DB devuelve una conexión limpia a PostgreSQL
func ObtenerConexion() (*sql.DB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
