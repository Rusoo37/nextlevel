package services

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"nextlevel/repository"
)

// IniciarLimpiadorTurnos arranca un hilo en segundo plano que revisa la DB periódicamente
func IniciarLimpiadorTurnos(db *sql.DB) {
	// Configuramos un reloj que haga "tic" cada 1 minuto
	ticker := time.NewTicker(1 * time.Minute)

	// Lanzamos la goroutine (el hilo ligero)
	go func() {
		for {
			// El código se frena acá hasta que el ticker avise que pasó 1 minuto
			<-ticker.C

			// Le decimos que limpie los que llevan más de 10 minutos pendientes
			filasCanceladas, err := repository.CancelarTurnosExpirados(db, 10)
			if err != nil {
				log.Printf("⚠️ Error en el limpiador automático: %v\n", err)
				continue // Si falla, que siga intentando en el próximo minuto
			}

			// Si encontró turnos colgados, nos avisa por consola
			if filasCanceladas > 0 {
				fmt.Printf("🧹 Limpieza automática: %d turno(s) expirado(s) liberado(s).\n", filasCanceladas)
			}
		}
	}()
}
