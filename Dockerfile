# 1. Usamos una imagen de Go para construir la app
FROM golang:1.24-alpine AS builder

# Instalamos git para descargar dependencias
RUN apk add --no-cache git

WORKDIR /app
COPY . .

# Compilamos el binario de forma estática
RUN go build -o main .

# 2. Usamos una imagen pequeña para correr el binario
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/static ./static

# Exponemos el puerto 8080
EXPOSE 8080

# Comando para ejecutar
CMD ["./main"]