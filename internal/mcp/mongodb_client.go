package mcp

import (
	"context"
	"fmt"
	"time"
)

// EjecutarAggregate llama a la herramienta 'aggregate' en el servidor MCP oficial de MongoDB
func EjecutarAggregate(ctx context.Context, db string, collection string, pipeline interface{}) (string, error) {
	if ServidorMCPMongo == nil {
		return "", fmt.Errorf("el servidor MCP de MongoDB no está inicializado")
	}

	arguments := map[string]interface{}{
		"db":         db,
		"collection": collection,
		"pipeline":   pipeline,
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	return ServidorMCPMongo.CallTool(ctxTimeout, "aggregate", arguments)
}
