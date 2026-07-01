package handlers

import (
	"fmt"
	"net/http"
	"time"
)

// IndexHandler maneja la página principal de la web de turnos
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html lang="es">
		<head>
			<meta charset="UTF-8">
			<title>NextLevel Necochea - Turnos</title>
		</head>
		<body>
			<h1>NextLevel Necochea</h1>
			<p>Sistema estructurado con buenas prácticas arquitectónicas.</p>
			<p>Fecha y hora del servidor: %s</p>
		</body>
		</html>
	`, time.Now().Format("02/01/2006 15:04:05"))
}
