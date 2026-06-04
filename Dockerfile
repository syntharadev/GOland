# Etapa 1: Builder
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copiar archivos de dependencias e instalarlas
COPY go.mod go.sum ./
RUN go mod download

# Copiar todo el código fuente
COPY . .

# Compilar AMBOS binarios
RUN CGO_ENABLED=0 go build -o /app/server ./cmd/server/main.go
RUN CGO_ENABLED=0 go build -o /app/mcp-mongo ./cmd/mcp-mongo/main.go

# Etapa 2: Runner
FROM alpine:latest

# Instalar certificados CA
RUN apk add --no-cache ca-certificates

WORKDIR /

# Copiar los binarios compilados al directorio raíz del Runner
COPY --from=builder /app/server ./server
COPY --from=builder /app/mcp-mongo ./mcp-mongo

# Copiar el directorio ui para servir los archivos frontend
COPY --from=builder /app/ui ./ui

# Exponer el puerto 8080
EXPOSE 8080

# Configurar el ENTRYPOINT
ENTRYPOINT ["./server"]
