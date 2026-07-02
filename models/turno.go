package models

import "time"

// Turno representa la estructura de un turno en el sistema
type Turno struct {
	ID              int       `json:"id"`
	FechaHoraInicio time.Time `json:"fecha_hora_inicio"`
	FechaHoraFin    time.Time `json:"fecha_hora_fin"`
	NombreCliente   string    `json:"nombre_cliente"`
	Telefono        string    `json:"telefono"`
	Estado          string    `json:"estado"` // PENDIENTE_PAGO, CONFIRMADO, CANCELADO, MANUAL
	MontoSenado     float64   `json:"monto_senado"`
	IDMercadoPago   string    `json:"id_mercadopago"`
	CreadoEn        time.Time `json:"creado_en"`
}

// Configuracion representa los ajustes que Ramón puede cambiar
type Configuracion struct {
	ID             int     `json:"id"`
	PrecioTurno    float64 `json:"precio_turno"`
	PorcentajeSena int     `json:"porcentaje_sena"`
}

// ExcepcionCalendario representa los días que Ramón no trabaja
type ExcepcionCalendario struct {
	ID     int       `json:"id"`
	Fecha  time.Time `json:"fecha"`
	Motivo string    `json:"motivo"`
}

type TurnoFijo struct {
	ID            int    `json:"id"`
	DiaSemana     int    `json:"dia_semana"`
	Hora          string `json:"hora"` // Formato "HH:MM"
	NombreCliente string `json:"nombre_cliente"`
	Activo        bool   `json:"activo"`
}

// SolicitudReserva es lo que recibimos del frontend cuando alguien quiere un turno
type SolicitudReserva struct {
	FechaHoraInicio string `json:"fecha_hora_inicio"`
	NombreCliente   string `json:"nombre_cliente"`
	Telefono        string `json:"telefono"`
}
