# Etapa 1: Builder
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copiar archivos de dependencias e instalarlas
COPY go.mod go.sum ./
RUN go mod download

# Copiar todo el código fuente
COPY . .

# Compilar los binarios
RUN CGO_ENABLED=0 go build -o /app/goland-server ./cmd/server
RUN CGO_ENABLED=0 go build -o /app/mcp-mongo-bin ./cmd/mcp-mongo

# Etapa 2: Runner (Runtime)
FROM alpine:latest

# Instalar certificados CA para poder consultar APIs externas de forma segura
RUN apk add --no-cache ca-certificates

WORKDIR /

# Copiar los binarios compilados desde el builder
COPY --from=builder /app/goland-server ./goland-server
COPY --from=builder /app/mcp-mongo-bin ./mcp-mongo-bin

# Copiar todos los archivos estáticos del frontend (incluyendo ui/html/app_GOland.html y ui/static)
COPY --from=builder /app/ui ./ui

# Exponer el puerto del servidor
EXPOSE 8080

# Definir el comando de arranque del servidor
CMD ["./goland-server"]
