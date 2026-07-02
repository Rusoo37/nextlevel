package config

import (
	"database/sql"

	_ "github.com/lib/pq"
)

// Ahora recibe la URL como parámetro
func ObtenerConexion(dbURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	// Verificamos que realmente conecte
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
