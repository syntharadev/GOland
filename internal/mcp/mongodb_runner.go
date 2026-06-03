package mcp

import (
	"fmt"
	"log"
)

// ServidorMCPMongo es la instancia global del cliente MCP conectado al servidor de MongoDB
var ServidorMCPMongo *MCPClient

// IniciarServidorMCPMongo arranca el servidor MCP nativo de MongoDB como subproceso
func IniciarServidorMCPMongo() error {
	log.Println("Iniciando Servidor MCP de MongoDB en Go como subproceso...")

	// Ejecutar nuestro propio binario/fuente MCP: go run cmd/mcp-mongo/main.go
	client, err := IniciarMCPClient("mongodb", "go", "run", "cmd/mcp-mongo/main.go")
	if err != nil {
		return fmt.Errorf("fallo al arrancar el servidor MCP nativo de MongoDB: %w", err)
	}

	ServidorMCPMongo = client
	log.Println("Servidor MCP nativo de MongoDB inicializado con éxito y enlazado.")
	return nil
}

// CerrarServidorMCPMongo detiene el servidor de MongoDB de forma ordenada
func CerrarServidorMCPMongo() {
	if ServidorMCPMongo != nil {
		ServidorMCPMongo.Close()
		ServidorMCPMongo = nil
	}
}
