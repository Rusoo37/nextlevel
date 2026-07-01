#!/bin/bash

# Colores para que se vea lindo en la terminal
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # Sin color

echo -e "${YELLOW}Iniciando batería de pruebas para la API de NextLevel...${NC}\n"

# ---------------------------------------------------------
echo -e "${YELLOW}Prueba 1: Falla intencional (Falta el nombre del cliente)${NC}"
curl -s -X POST http://localhost:8080/api/reservar \
-H "Content-Type: application/json" \
-w "\nCódigo HTTP: %{http_code}\n" \
-d '{
  "fecha_hora_inicio": "2026-07-18T10:00:00Z", 
  "nombre_cliente": "", 
  "telefono": "2262123456", 
  "email": "prueba@test.com"
}'
echo -e "---------------------------------------------------------\n"

# ---------------------------------------------------------
echo -e "${GREEN}Prueba 2: Éxito (Reserva válida y Link de Pago)${NC}"
curl -s -X POST http://localhost:8080/api/reservar \
-H "Content-Type: application/json" \
-w "\nCódigo HTTP: %{http_code}\n" \
-d '{
  "fecha_hora_inicio": "2026-07-01T17:00:00Z", 
  "nombre_cliente": "Cliente maldito", 
  "telefono": "2262112233", 
  "email": "auto@test.com"
}'
echo -e "---------------------------------------------------------\n"

# ---------------------------------------------------------
echo -e "${RED}Prueba 3: Falla intencional (Concurrencia - Mismo horario)${NC}"
echo "(Intentando robar el turno de las 10:00 que acabamos de reservar...)"
curl -s -X POST http://localhost:8080/api/reservar \
-H "Content-Type: application/json" \
-w "\nCódigo HTTP: %{http_code}\n" \
-d '{
  "fecha_hora_inicio": "2026-07-16T10:00:00Z", 
  "nombre_cliente": "El Ladrón de Turnos", 
  "telefono": "2262998877", 
  "email": "ladron@test.com"
}'
echo -e "---------------------------------------------------------\n"

# ---------------------------------------------------------
echo -e "${GREEN}Prueba 4: Consultar Disponibilidad${NC}"
echo "(Las 10:00 AM ya no debería aparecer en la lista)"
curl -s -X GET "http://localhost:8080/api/disponibilidad?fecha=2026-07-16" \
-w "\nCódigo HTTP: %{http_code}\n"
echo -e "\n---------------------------------------------------------\n"

echo -e "${YELLOW}Fin de las pruebas.${NC}"