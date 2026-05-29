# ETAPA 1: Construcción (Builder)
FROM golang:1.26.3-alpine AS builder

# Configurar el directorio de trabajo
WORKDIR /app

# Descargar dependencias de Go
COPY go.mod go.sum ./
RUN go mod download

# Copiar el código fuente completo (¡Todo en una sola línea!)
COPY . .

# Compilar el binario estático sin dependencias C (CGO_ENABLED=0)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o goland-server ./cmd/server/main.go

# ETAPA 2: Ejecución (Runner) - Imagen hiperligera
FROM alpine:latest

# Añadir certificados raíz para que el servidor pueda conectarse a la API de Gemini vía HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copiar el binario compilado desde la etapa builder
COPY --from=builder /app/goland-server .
# Copiar la carpeta estática del frontend (Agentic UI)
COPY --from=builder /app/ui ./ui

# Exponer el puerto
EXPOSE 8080

# Comando de arranque
CMD ["./goland-server"]
