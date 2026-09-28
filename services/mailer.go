package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"time"
)

type DatosTurno struct {
	Nombre string
	Email  string
	Fecha  string // ej: "04/10/2026"
	Hora   string // ej: "15:30"
	Seña   string // ej: "$5000"
}

var httpClientMail = &http.Client{Timeout: 10 * time.Second}

// EnviarMailTurno manda el mail de confirmación usando la API transaccional de Brevo.
func EnviarMailTurno(t DatosTurno) error {
	apiKey := os.Getenv("BREVO_API_KEY")
	fromEmail := os.Getenv("MAIL_FROM_EMAIL")
	fromName := os.Getenv("MAIL_FROM_NAME")
	if apiKey == "" || fromEmail == "" {
		return fmt.Errorf("faltan BREVO_API_KEY o MAIL_FROM_EMAIL")
	}
	if fromName == "" {
		fromName = "NextLevel Necochea"
	}

	cuerpo := fmt.Sprintf(`
		<h2>¡Tu turno está confirmado!</h2>
		<p>Hola %s, recibimos tu seña de %s.</p>
		<p><b>Día:</b> %s<br><b>Hora:</b> %s</p>
		<p>Si necesitás cancelar o reprogramar, contactanos con anticipación.</p>`,
		html.EscapeString(t.Nombre), html.EscapeString(t.Seña),
		html.EscapeString(t.Fecha), html.EscapeString(t.Hora))

	campos := map[string]any{
		"sender":      map[string]string{"name": fromName, "email": fromEmail},
		"to":          []map[string]string{{"email": t.Email, "name": t.Nombre}},
		"subject":     "Confirmación de tu turno",
		"htmlContent": cuerpo,
	}
	if replyTo := os.Getenv("MAIL_REPLY_TO"); replyTo != "" {
		campos["replyTo"] = map[string]string{"email": replyTo}
	}

	payload, err := json.Marshal(campos)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.brevo.com/v3/smtp/email", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("api-key", apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := httpClientMail.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		detalle, _ := io.ReadAll(io.LimitReader(resp.Body, 500))
		return fmt.Errorf("brevo respondió %d: %s", resp.StatusCode, detalle)
	}
	return nil
}
