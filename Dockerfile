# Etapa 1: Builder
FROM golang:1.26-alpine AS builder

# Instalar herramientas básicas de compilación
RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# Copiar archivos de módulos
COPY go.mod go.sum ./
RUN go mod download

# Copiar todo el código fuente
COPY . .

# Compilar de forma estática ambos binarios para Linux Alpine
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/mcp-mongo ./cmd/mcp-mongo/main.go

# Etapa 2: Runner
FROM alpine:latest

# Instalar certificados CA y zona horaria
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /

# Copiar los binarios compilados
COPY --from=builder /app/server /server
COPY --from=builder /app/mcp-mongo /mcp-mongo

# Copiar el directorio ui para servir archivos estáticos, videos e HTML
COPY --from=builder /app/ui /ui

# Exponer el puerto del servidor web
EXPOSE 8080

# Definir el comando de inicio
ENTRYPOINT ["/server"]
