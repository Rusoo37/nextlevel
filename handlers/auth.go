package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

// AuthMiddleware es el "patovica". Envuelve las rutas y verifica la cookie.
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("admin_session")

		// Si no tiene la cookie o la pulsera es trucha, lo mandamos al login
		if err != nil || cookie.Value != "ramon_autorizado" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Si está todo bien, lo dejamos pasar a la ruta original
		next(w, r)
	}
}

// LoginVistaHandler muestra la pantalla visual para poner la clave
func LoginVistaHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/html/login.html")
}

// LoginProcesoHandler recibe el intento de login y decide si lo aprueba
func LoginProcesoHandler(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Usuario  string `json:"usuario"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Datos inválidos", http.StatusBadRequest)
		return
	}

	// hay que pasarla a una variable de entorno
	if creds.Usuario == "ramon" && creds.Password == "peluqueria123" {

		// coockie por 24 horas, para que no tenga que loguearse cada vez que recarga la página
		vencimiento := time.Now().Add(24 * time.Hour)
		cookie := &http.Cookie{
			Name:     "admin_session",
			Value:    "ramon_autorizado",
			Expires:  vencimiento,
			HttpOnly: true, // esto hace que no se pueda acceder a la cookie desde JS, solo desde el servidor
			Path:     "/",
		}

		http.SetCookie(w, cookie)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"mensaje": "Éxito"})
		return
	}

	http.Error(w, "Usuario o contraseña incorrectos", http.StatusUnauthorized)
}
