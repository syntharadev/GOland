package agents

import (
	"context"
	"fmt"
	"log"
	"strings"

	"gemini-go-platform/internal/llm"
	"gemini-go-platform/internal/mcp"

	"github.com/google/generative-ai-go/genai"
)

// BibliotecariaAgent orquesta la interacción con Gemini y la llamada a herramientas de agregación de MongoDB
type BibliotecariaAgent struct {
	Gemini *llm.GeminiClient
}

// NewBibliotecariaAgent inicializa un nuevo agente de La Bibliotecaria
func NewBibliotecariaAgent(client *llm.GeminiClient) *BibliotecariaAgent {
	return &BibliotecariaAgent{
		Gemini: client,
	}
}

// ConsultarDocumentacion ejecuta el flujo completo de RAG mediante búsqueda de expresiones regulares en MongoDB
func (b *BibliotecariaAgent) ConsultarDocumentacion(ctx context.Context, query string) (string, error) {
	systemInstruction := "Eres La Bibliotecaria del ecosistema GOland. Tu objetivo es ayudar a los estudiantes resolviendo sus dudas de programación en Go. Responde de forma clara y didáctica basándote estrictamente en el contexto proporcionado. Si el contexto está vacío o no es relevante, indícalo de forma educada y responde según tus conocimientos generales de Go, aclarando que no se encontró documentación oficial al respecto en el RAG."

	// 1. Caso de simulación o mock si el cliente no está inicializado
	if b.Gemini == nil || b.Gemini.Client == nil {
		log.Println("Bibliotecaria [MOCK]: Cliente Gemini no configurado. Simulando consulta en MongoDB...")
		return fmt.Sprintf("¡Hola! Soy La Bibliotecaria de GOland (Modo Simulación). He consultado el servidor MCP de MongoDB Atlas sobre '%s' y recuperado las guías oficiales de concurrencia y canales. ¡El RAG funciona!", query), nil
	}

	// 2. Extraer palabras clave para la regex
	words := strings.Fields(query)
	var keywords []string
	for _, w := range words {
		wClean := strings.Trim(w, ",.?/!@#$^*()_+-=")
		if len(wClean) > 2 {
			keywords = append(keywords, wClean)
		}
	}

	var regexPattern string
	if len(keywords) > 0 {
		regexPattern = "(?i)" + strings.Join(keywords, "|")
	} else {
		regexPattern = "(?i)" + query
	}

	log.Printf("Bibliotecaria [REAL]: Ejecutando búsqueda RAG por Regex con patrón: '%s'", regexPattern)

	// 3. Construir el pipeline de agregación para MongoDB
	pipeline := []interface{}{
		map[string]interface{}{
			"$match": map[string]interface{}{
				"content": map[string]interface{}{
					"$regex": regexPattern,
				},
			},
		},
		map[string]interface{}{
			"$limit": 3,
		},
	}

	// 4. Ejecutar la consulta de agregación en MongoDB MCP
	resultadoAgregacion, err := mcp.EjecutarAggregate(ctx, "goland_db", "go_docs", pipeline)
	if err != nil {
		log.Printf("Error al consultar documentación mediante Regex en MongoDB: %v. Continuando sin contexto...", err)
		resultadoAgregacion = "No se pudo recuperar contexto de la base de datos debido a un error en el RAG."
	}

	// 5. Enviar a Gemini para la síntesis de la respuesta
	model := b.Gemini.Client.GenerativeModel("gemini-2.5-flash")
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(systemInstruction)},
	}

	promptContexto := fmt.Sprintf("Contexto recuperado de MongoDB (RAG):\n%s\n\nPregunta del usuario:\n%s", resultadoAgregacion, query)

	respFinal, err := model.GenerateContent(ctx, genai.Text(promptContexto))
	if err != nil {
		return "", fmt.Errorf("error al sintetizar respuesta RAG con Gemini: %w", err)
	}

	if len(respFinal.Candidates) == 0 || respFinal.Candidates[0].Content == nil || len(respFinal.Candidates[0].Content.Parts) == 0 {
		return "No se pudo sintetizar la respuesta final de La Bibliotecaria.", nil
	}

	finalPart := respFinal.Candidates[0].Content.Parts[0]
	if textFinal, isText := finalPart.(genai.Text); isText {
		return string(textFinal), nil
	}

	return "No se pudo decodificar el veredicto de La Bibliotecaria.", nil
}
