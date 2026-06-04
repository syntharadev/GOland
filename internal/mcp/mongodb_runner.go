package mcp

import (
	"fmt"
	"log"
	"os"
)

// ServidorMCPMongo es la instancia global del cliente MCP conectado al servidor de MongoDB
var ServidorMCPMongo *MCPClient

// IniciarServidorMCPMongo arranca el servidor MCP nativo de MongoDB como subproceso
func IniciarServidorMCPMongo() error {
	var client *MCPClient
	var err error

	if os.Getenv("ENV") == "production" {
		log.Println("Iniciando Servidor MCP de MongoDB en Go (Modo Producción - Binario compilado)...")
		client, err = IniciarMCPClient("mongodb", "./mcp-mongo-bin")
	} else {
		log.Println("Iniciando Servidor MCP de MongoDB en Go (Modo Desarrollo - go run)...")
		client, err = IniciarMCPClient("mongodb", "go", "run", "cmd/mcp-mongo/main.go")
	}

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
