package services

import (
	"context"
	"fmt"
	"os"

	// Usamos las rutas correctas del SDK de Go
	"github.com/mercadopago/sdk-go/pkg/config"
	"github.com/mercadopago/sdk-go/pkg/preference"
)

// GenerarLinkDePago crea la preferencia en Mercado Pago y devuelve la URL para cobrar
func GenerarLinkDePago(idTurno int, monto float64, nombreCliente string) (string, error) {
	// 1. Autenticación

	accessToken := os.Getenv("MP_ACCESS_TOKEN")
	cfg, err := config.New(accessToken)
	if err != nil {
		return "", fmt.Errorf("error configurando mercadopago: %v", err)
	}

	// 2. Armamos el cliente de preferencias
	client := preference.NewClient(cfg)

	// 3. Lógica dinámica para la URL del Webhook
	webhookURL := os.Getenv("WEBHOOK_URL")
	if webhookURL == "" {
		// Si no hay variable configurada (estás en tu compu), usamos Ngrok
		webhookURL = "https://perm-cried-papyrus.ngrok-free.dev/api/webhook"
	} else {
		// Si hay variable (ej: estás en Render), le pegamos la ruta al final
		webhookURL = webhookURL + "/api/webhook"
	}

	// 4. Definimos qué le estamos cobrando
	request := preference.Request{
		Items: []preference.ItemRequest{
			{
				Title:       "Seña Turno Peluquería - Ramón",
				Description: fmt.Sprintf("Turno para %s", nombreCliente),
				Quantity:    1,
				UnitPrice:   monto,
				CurrencyID:  "ARS",
			},
		},
		BackURLs: &preference.BackURLsRequest{
			// Nota: Cuando subamos a producción, estas URLs también las haremos dinámicas,
			// pero por ahora está perfecto que apunten a tu dominio real.
			Success: "https://nextlevel-r3jg.onrender.com/exito",
			Failure: "https://nextlevel_r3jg.onrender.com/fallo",
			Pending: "https://nextlevel_r3jg.onrender.com/pendiente",
		},
		AutoReturn:        "approved",
		ExternalReference: fmt.Sprintf("%d", idTurno),

		// Usamos la variable dinámica que calculamos más arriba
		NotificationURL: webhookURL,
	}

	// 5. Se lo mandamos a Mercado Pago
	resource, err := client.Create(context.Background(), request)
	if err != nil {
		return "", fmt.Errorf("error creando preferencia: %v", err)
	}

	// Devolvemos el link donde el usuario tiene que pagar
	return resource.InitPoint, nil
}
