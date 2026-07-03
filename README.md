# NextLevel - Sistema de Reserva de Turnos y Gestión para Peluquería

NextLevel es una solución web integral diseñada para automatizar la reserva de turnos en peluquerías y barberías, eliminando el problema de los turnos perdidos y cancelaciones de último momento mediante la integración de un sistema dinámico de señas obligatorias.

Desplegado en producción y utilizado en un entorno real.

## 🚀 Características del Proyecto

- **Reserva de Turnos en Tiempo Real:** Interfaz limpia en el cliente que permite visualizar los bloques de horarios disponibles por día.
- **Bloqueos Dinámicos:** Exclusión automática de turnos pasados del día corriente.
- **Turnos Fijos/Recurrentes:** Panel administrativo capaz de bloquear horarios fijos semanales para clientes habituales.
- **Automatización de Pagos (Mercado Pago):** Integración nativa con la API de Mercado Pago para procesar cobros de señas temporales.
- **Confirmación por Webhook:** Flujo asincrónico que procesa notificaciones de pago seguras directas de Mercado Pago, pasando el turno de `PENDIENTE` a `CONFIRMADO` sin intervención manual.
- **Limpieza Automática (Cron Job):** Hilo de ejecución en segundo plano (Go goroutine) que libera automáticamente los turnos pre-reservados si el pago no se completa en un tiempo límite.

## 🛠️ Stack Tecnológico

- **Backend:** Go (Golang) 1.25+ utilizando exclusivamente la librería estándar para routing de alto rendimiento.
- **Base de Datos:** PostgreSQL alojado en la nube con **Neon**.
- **Frontend:** HTML5, CSS3, JavaScript - Arquitectura SPA minimalista optimizada para móviles.
- **Contenerización y Deploy:** Docker (Multi-stage builds) desplegado en **Render** con variables de entorno protegidas.
- **Integraciones:** Mercado Pago API (SDK Go).

## 📐 Arquitectura de la Solución

El flujo del sistema garantiza la integridad del calendario incluso ante alta concurrencia:

1. **Selección:** El cliente elige fecha/hora -> El backend valida la disponibilidad en base horaria local.
2. **Pre-reserva:** Se crea un registro temporal en la base de datos bloqueando el espacio por $X$ minutos.
3. **Pasarela:** El backend genera una preferencia en Mercado Pago inyectando el ID del turno en el campo `external_reference`.
4. **Verificación (Conciliación):** Mercado Pago dispara un Webhook al servidor. El handler valida el ID del pago mediante una consulta directa (GET) a la API de Mercado Pago por seguridad y actualiza el estado del turno en la DB.
