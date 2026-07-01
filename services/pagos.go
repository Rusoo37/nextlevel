package services

import (
	"context"
	"fmt"

	// Usamos las rutas correctas del SDK de Go
	"github.com/mercadopago/sdk-go/pkg/config"
	"github.com/mercadopago/sdk-go/pkg/preference"
)

// GenerarLinkDePago crea la preferencia en Mercado Pago y devuelve la URL para cobrar
func GenerarLinkDePago(idTurno int, monto float64, nombreCliente string) (string, error) {
	// 1. Autenticación
	accessToken := "APP_USR-4638107181664481-070112-eaf8d20647e2a3813e36356163ae22ac-3509855779"

	cfg, err := config.New(accessToken)
	if err != nil {
		return "", fmt.Errorf("error configurando mercadopago: %v", err)
	}

	// 2. Armamos el cliente de preferencias
	client := preference.NewClient(cfg)

	// 3. Definimos qué le estamos cobrando
	// 3. Definimos qué le estamos cobrando
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
			Success: "https://nextlevel_necochea.com.ar/exito",
			Failure: "https://nextlevel_necochea.com.ar/fallo",
			Pending: "https://nextlevel_necochea.com.ar/pendiente",
		},
		AutoReturn:        "approved",
		ExternalReference: fmt.Sprintf("%d", idTurno),

		// ---------------------------------------------------------
		// 🔥 LA CLAVE: Forzamos a MP a que avise acá
		// ---------------------------------------------------------
		NotificationURL: "https://perm-cried-papyrus.ngrok-free.dev/api/webhook",
	}

	// 4. Se lo mandamos a Mercado Pago
	resource, err := client.Create(context.Background(), request)
	if err != nil {
		return "", fmt.Errorf("error creando preferencia: %v", err)
	}

	// Devolvemos el link donde el usuario tiene que pagar (InitPoint)
	return resource.InitPoint, nil
}
