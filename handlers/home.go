package handlers

import (
	"net/http"
)

// IndexHandler sirve la página principal de la web de turnos
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Le decimos a Go que sirva el archivo index.html físico
	http.ServeFile(w, r, "./static/html/index.html")
}

// ExitoHandler sirve la página de pago aprobado
func ExitoHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/html/exito.html")
}

// FalloHandler sirve la página de pago rechazado
func FalloHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/html/fallo.html")
}

// PendienteHandler sirve la página de pago en proceso
func PendienteHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/html/pendiente.html")
}

// AdminHandler sirve la página del panel de control
func AdminHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/html/admin.html")
}
