package mcp

import (
	"fmt"
	"log"

	"gemini-go-platform/internal/config"
)

// ServidorMCPMongo es la instancia global del cliente MCP conectado al servidor de MongoDB
var ServidorMCPMongo *MCPClient

// IniciarServidorMCPMongo arranca el servidor MCP oficial de MongoDB como subproceso
func IniciarServidorMCPMongo() error {
	mongodbURI := config.GetMongoDBURI()
	log.Printf("Iniciando Servidor MCP MongoDB con URI: %s", mongodbURI)

	// Ejecutar npx -y @modelcontextprotocol/server-mongodb <MONGODB_URI>
	client, err := IniciarMCPClient("mongodb", "npx", "-y", "@modelcontextprotocol/server-mongodb", mongodbURI)
	if err != nil {
		return fmt.Errorf("fallo al arrancar el servidor MCP MongoDB: %w", err)
	}

	ServidorMCPMongo = client
	log.Println("Servidor MCP de MongoDB Atlas inicializado con éxito y enlazado.")
	return nil
}

// CerrarServidorMCPMongo detiene el servidor de MongoDB de forma ordenada
func CerrarServidorMCPMongo() {
	if ServidorMCPMongo != nil {
		ServidorMCPMongo.Close()
		ServidorMCPMongo = nil
	}
}
