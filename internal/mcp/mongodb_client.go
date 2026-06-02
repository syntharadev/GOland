package mcp

import (
	"encoding/json"
	"fmt"
	"strings"
)

// DocumentoDoc representa el formato del resultado obtenido de MongoDB Atlas Vector Search
type DocumentoDoc struct {
	Titulo      string  `json:"titulo"`
	Contenido   string  `json:"contenido"`
	Fuente      string  `json:"fuente"`
	Score       float64 `json:"score"`
}

// EjecutarBusquedaVectorial simula una consulta RAG al servidor MCP de MongoDB Atlas Vector Search
func EjecutarBusquedaVectorial(query string) (string, error) {
	queryLower := strings.ToLower(query)

	var resultados []DocumentoDoc

	if strings.Contains(queryLower, "goroutine") || strings.Contains(queryLower, "concurrencia") {
		resultados = append(resultados, DocumentoDoc{
			Titulo:    "Fundamentos de Concurrencia: Goroutines",
			Contenido: "Una goroutine es un hilo de ejecución ligero administrado por el runtime de Go. Sintaxis: `go f(x, y, z)`. Las goroutines comparten el mismo espacio de direcciones, por lo que el acceso a la memoria compartida debe ser sincronizado utilizando canales o el paquete `sync`.",
			Fuente:    "https://go.dev/tour/concurrency/1",
			Score:     0.98,
		})
	}

	if strings.Contains(queryLower, "canal") || strings.Contains(queryLower, "channel") || strings.Contains(queryLower, "concurrencia") {
		resultados = append(resultados, DocumentoDoc{
			Titulo:    "Canales de Comunicación (Channels)",
			Contenido: "Los canales son los conductos tipados a través de los cuales puedes enviar y recibir valores con el operador de canal, `<-`. Se inicializan usando `make(chan int)`. Por defecto, los envíos y recepciones se bloquean hasta que el otro lado esté listo, lo que permite sincronizar goroutines sin bloqueos explícitos.",
			Fuente:    "https://go.dev/tour/concurrency/2",
			Score:     0.95,
		})
	}

	if strings.Contains(queryLower, "variable") || strings.Contains(queryLower, "declara") || strings.Contains(queryLower, "sintaxis") {
		resultados = append(resultados, DocumentoDoc{
			Titulo:    "Variables e Inferencia de Tipos",
			Contenido: "En Go, las variables se declaran explícitamente usando la palabra clave `var` (ej. `var x int`). Dentro de una función, la asignación corta de declaración `:=` puede ser usada para declarar e inicializar una variable con inferencia implícita de tipo (ej. `y := 42`).",
			Fuente:    "https://go.dev/tour/basics/10",
			Score:     0.92,
		})
	}

	// Fallback genérico si no se encuentran palabras clave específicas
	if len(resultados) == 0 {
		resultados = append(resultados, DocumentoDoc{
			Titulo:    "Documentación General de GOland",
			Contenido: "GOland es un ecosistema educativo adaptativo para dominar el desarrollo en lenguaje Go. Admite la inicialización interactiva de entornos y la compilación directa. El tutor Gemini puede ser consultado en cualquier momento para obtener explicaciones y validar conceptos del lenguaje.",
			Fuente:    "https://goland.dev/docs/intro",
			Score:     0.80,
		})
	}

	// Convertir los resultados a formato JSON para retornar al agente
	jsonData, err := json.Marshal(resultados)
	if err != nil {
		return "", fmt.Errorf("error serializando resultados MCP: %w", err)
	}

	return string(jsonData), nil
}
