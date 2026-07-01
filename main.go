package main

import (
	"fmt"
	"log"
	"net/http"

	"nextlevel/config"
	"nextlevel/handlers"
	"nextlevel/repository"
	"nextlevel/services"
)

func main() {
	// 1. Obtener conexión a la DB a través del paquete config
	db, err := config.ObtenerConexion()
	if err != nil {
		log.Fatalf("Error crítico al conectar la DB: %v", err)
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

	http.HandleFunc("/", handlers.IndexHandler)

	// Instanciamos el handler de la API pasándole la base de datos
	api := &handlers.APIHandler{DB: db}
	http.HandleFunc("/api/disponibilidad", api.Disponibilidad)
	http.HandleFunc("/api/reservar", api.Reservar)

	// 4. Lanzar servidor
	fmt.Println("Servidor corriendo en http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
