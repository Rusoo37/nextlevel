package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"nextlevel/config"
	"nextlevel/handlers"
	"nextlevel/repository"
	"nextlevel/services"
	_ "time/tzdata"

	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println("Error cargando .env: asegúrate de que el archivo exista en la raíz del proyecto")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("La variable DB_URL no está configurada en el archivo .env")
	}

	db, err := config.ObtenerConexion(dbURL)
	if err != nil {
		log.Fatalf("Error al conectar a la DB: %v", err)
	}
	defer db.Close()
	fmt.Println("¡Conexión a PostgreSQL exitosa!")

	// 2. Inicializar tablas a través del paquete repository
	if err := repository.InicializarTablas(db); err != nil {
		log.Fatalf("Error crítico al inicializar tablas: %v", err)
	}
	fmt.Println("Estructura de base de datos verificada de manera limpia.")

	services.IniciarLimpiadorTurnos(db)
	fmt.Println("Limpiador de turnos automático corriendo en segundo plano...")

	// 3. Registrar rutas usando los handlers dedicados
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/login", handlers.LoginVistaHandler)
	http.HandleFunc("/api/login", handlers.LoginProcesoHandler)

	http.HandleFunc("/", handlers.IndexHandler)
	http.HandleFunc("/exito", handlers.ExitoHandler)
	http.HandleFunc("/fallo", handlers.FalloHandler)
	http.HandleFunc("/pendiente", handlers.PendienteHandler)

	// Instanciamos el handler de la API pasándole la base de datos
	api := &handlers.APIHandler{DB: db}
	http.HandleFunc("/api/disponibilidad", api.Disponibilidad)
	http.HandleFunc("/api/reservar", api.Reservar)
	http.HandleFunc("/api/webhook", api.WebhookMercadoPago)
	http.HandleFunc("/api/turno", api.DetalleTurno)

	// rutas de admin con middleware de autenticación
	http.HandleFunc("/admin", handlers.AuthMiddleware(handlers.AdminHandler))
	http.HandleFunc("/api/admin/turnos", handlers.AuthMiddleware(api.AdminTurnosHandler))
	http.HandleFunc("/api/admin/bloquear", handlers.AuthMiddleware(api.BloquearHorarioAdmin))

	http.HandleFunc("/api/admin/config/fijos", handlers.AuthMiddleware(api.ConfigFijosHandler))
	http.HandleFunc("/api/admin/config/precios", handlers.AuthMiddleware(api.ConfigPreciosHandler))
	http.HandleFunc("/api/admin/config/excepciones", handlers.AuthMiddleware(api.ConfigExcepcionesHandler))

	// 4. Lanzar servidor
	fmt.Println("Servidor corriendo en http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
