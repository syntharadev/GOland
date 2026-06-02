package mcp

import (
	"encoding/json"
	"fmt"
	"regexp"
)

// ElasticTelemetry representa los datos indexados en ElasticSearch APM
type ElasticTelemetry struct {
	Complexity    string `json:"complexity"`     // Complejidad Big O (O(1), O(n), O(n^2))
	Memory        string `json:"memory"`         // Memoria consumida
	ExecutionTime string `json:"execution_time"` // Tiempo de ejecución estimado
	Status        string `json:"status"`
}

// IndexarMetricasRendimiento analiza el código e indexa métricas de APM simuladas en ElasticSearch
func IndexarMetricasRendimiento(codigo string) (string, error) {
	// Expresión regular para detectar bucles for anidados
	nestedForRegex := regexp.MustCompile(`for\s+.*\{\s*[\s\S]*?for\s+.*\{`)
	isNested := nestedForRegex.MatchString(codigo)

	var telemetry ElasticTelemetry
	telemetry.Status = "indexed"

	if isNested {
		telemetry.Complexity = "O(n^2)"
		telemetry.Memory = "Alto (12MB)"
		telemetry.ExecutionTime = "45ms"
	} else {
		// Heurística para O(n) o O(1)
		hasFor := regexp.MustCompile(`for\s+.*\{`).MatchString(codigo)
		if hasFor {
			telemetry.Complexity = "O(n)"
			telemetry.Memory = "Óptima (2MB)"
			telemetry.ExecutionTime = "1.2ms"
		} else {
			telemetry.Complexity = "O(1)"
			telemetry.Memory = "Mínima (400KB)"
			telemetry.ExecutionTime = "0.08ms"
		}
	}

	jsonData, err := json.Marshal(telemetry)
	if err != nil {
		return "", fmt.Errorf("error serializando métricas de Elastic MCP: %w", err)
	}

	return string(jsonData), nil
}
