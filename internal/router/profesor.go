package router

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"gemini-go-platform/internal/llm"
	"github.com/google/generative-ai-go/genai"
)

type IntencionResponse struct {
	Escuadron int    `json:"escuadron"`
	Razon     string `json:"razon"`
}

// ClasificarIntencion determina a qué escuadrón enviar la consulta del estudiante
func ClasificarIntencion(ctx context.Context, gemini *llm.GeminiClient, input string, codigo string) (int, string, error) {
	systemPrompt := `Eres "El Profesor", el nodo central de enrutamiento de GOland. Tu tarea es analizar la consulta del estudiante y clasificarla en uno de los 4 escuadrones especializados (Swarm).
Devuelve ÚNICAMENTE un objeto JSON con el siguiente formato:
{"escuadron": X, "razon": "breve explicación de la clasificación"}
donde X es un número entero del 1 al 4 según las siguientes reglas:
1: Teoría / Dudas / Conceptos / Explicaciones generales sobre Go (se asigna a Swarm 1: Bibliotecaria, Hacker)
2: Arquitectura / Estructura del proyecto / Flujo de archivos (se asigna a Swarm 2: Ingeniera, Constructor)
3: Sintaxis / Errores de compilación / Logs de consola (se asigna a Swarm 3: Senior, Mensajero)
4: Optimización / Algoritmos / Complejidad / Tests (se asigna a Swarm 4: Guardiana, Cronometradora)

No incluyas explicaciones adicionales, ni markdown, ni formato de código como ` + "`" + `json` + "`" + `. Solo el JSON puro.`

	if gemini == nil || gemini.Client == nil {
		log.Println("Profesor [MOCK]: Cliente Gemini no configurado. Simulando clasificación...")
		// Clasificación simulada simple
		escuadron := 1
		razon := "Consulta sobre teoría o general de Go (modo simulación)."
		inputLower := strings.ToLower(input)
		if strings.Contains(inputLower, "error") || strings.Contains(inputLower, "compila") || strings.Contains(inputLower, "sintaxis") {
			escuadron = 3
			razon = "Detectadas palabras clave de sintaxis o error."
		} else if strings.Contains(inputLower, "test") || strings.Contains(inputLower, "optimiza") || strings.Contains(inputLower, "eficiencia") {
			escuadron = 4
			razon = "Detectadas palabras clave de optimización o tests."
		} else if strings.Contains(inputLower, "arquitectura") || strings.Contains(inputLower, "estructura") || strings.Contains(inputLower, "crear") {
			escuadron = 2
			razon = "Detectadas palabras clave de arquitectura o estructura."
		}
		return escuadron, razon, nil
	}

	model := gemini.Client.GenerativeModel("gemini-2.5-flash")
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(systemPrompt)},
	}

	prompt := fmt.Sprintf("Consulta del estudiante: %s\nCódigo opcional: %s", input, codigo)
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return 1, "Error de llamada a Gemini", err
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return 1, "Respuesta vacía de Gemini", fmt.Errorf("no candidates in response")
	}

	textPart, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return 1, "Respuesta no es de tipo texto", fmt.Errorf("part is not text")
	}

	cleanJSON := string(textPart)
	// Limpiar posibles bloques de código markdown
	cleanJSON = strings.TrimPrefix(cleanJSON, "```json")
	cleanJSON = strings.TrimPrefix(cleanJSON, "```")
	cleanJSON = strings.TrimSuffix(cleanJSON, "```")
	cleanJSON = strings.TrimSpace(cleanJSON)

	var res IntencionResponse
	if err := json.Unmarshal([]byte(cleanJSON), &res); err != nil {
		log.Printf("Aviso: Fallo al deserializar JSON de ClasificarIntencion: %v. Contenido: %q", err, cleanJSON)
		// Fallback inteligente
		return 1, "Fallo al decodificar clasificación (fallback a Swarm 1)", nil
	}

	if res.Escuadron < 1 || res.Escuadron > 4 {
		res.Escuadron = 1
	}

	return res.Escuadron, res.Razon, nil
}
