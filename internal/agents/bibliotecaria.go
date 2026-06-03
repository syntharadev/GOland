package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

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

// ConsultarDocumentacion ejecuta el flujo completo de RAG con Tool Calling en MongoDB Atlas
func (b *BibliotecariaAgent) ConsultarDocumentacion(ctx context.Context, query string) (string, error) {
	systemInstruction := "Eres La Bibliotecaria del ecosistema GOland. Tu única herramienta es 'aggregate' para buscar en MongoDB. Nunca inventes código. Responde basándote estrictamente en el contexto devuelto por MongoDB a través de tu etapa $vectorSearch. Si no hay contexto relevante, indícalo de forma educada."

	// 1. Caso de simulación o mock si el cliente no está inicializado
	if b.Gemini == nil || b.Gemini.Client == nil {
		log.Println("Bibliotecaria [MOCK]: Cliente Gemini no configurado. Simulando consulta en MongoDB...")
		return fmt.Sprintf("¡Hola! Soy La Bibliotecaria de GOland (Modo Simulación). He consultado el servidor MCP de MongoDB Atlas sobre '%s' y recuperado las guías oficiales de concurrencia y canales. ¡El RAG funciona!", query), nil
	}

	// 2. Flujo real con la API de Gemini
	model := b.Gemini.Client.GenerativeModel("gemini-2.5-flash")
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(systemInstruction)},
	}

	// Declaración de la herramienta 'aggregate' oficial del MCP de MongoDB
	funcDecl := &genai.FunctionDeclaration{
		Name:        "aggregate",
		Description: "Ejecuta un pipeline de agregación en la base de datos de MongoDB. Se usa para realizar búsquedas vectoriales sobre la documentación con la etapa $vectorSearch (usando 'queryVector': [0] como marcador de posición).",
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"db": {
					Type:        genai.TypeString,
					Description: "El nombre de la base de datos (usar siempre 'goland_db').",
				},
				"collection": {
					Type:        genai.TypeString,
					Description: "La colección de destino (usar siempre 'go_docs').",
				},
				"pipeline": {
					Type:        genai.TypeString,
					Description: "El array del pipeline de agregación JSON, conteniendo la etapa '$vectorSearch' con un array numérico de relleno en 'queryVector'.",
				},
			},
			Required: []string{"db", "collection", "pipeline"},
		},
	}

	model.Tools = []*genai.Tool{
		{
			FunctionDeclarations: []*genai.FunctionDeclaration{funcDecl},
		},
	}

	// Enviar pregunta del estudiante a Gemini
	resp, err := model.GenerateContent(ctx, genai.Text(query))
	if err != nil {
		return "", fmt.Errorf("error en llamada inicial a Gemini en La Bibliotecaria: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return "No se ha obtenido respuesta del modelo de razonamiento.", nil
	}

	part := resp.Candidates[0].Content.Parts[0]

	// Verificar si Gemini invoca la herramienta 'aggregate'
	funcCall, ok := part.(genai.FunctionCall)
	if !ok {
		if textPart, isText := part.(genai.Text); isText {
			return string(textPart), nil
		}
		return "Respuesta sin contenido estructurado de La Bibliotecaria.", nil
	}

	if funcCall.Name != "aggregate" {
		return "", fmt.Errorf("Gemini solicitó una herramienta inesperada: %s", funcCall.Name)
	}

	// Extraer argumentos
	dbVal, _ := funcCall.Args["db"].(string)
	collectionVal, _ := funcCall.Args["collection"].(string)
	pipelineStr, _ := funcCall.Args["pipeline"].(string)

	if dbVal == "" {
		dbVal = "goland_db"
	}
	if collectionVal == "" {
		collectionVal = "go_docs"
	}

	log.Printf("Bibliotecaria [REAL]: Gemini solicitó agregación en db: '%s', coll: '%s'", dbVal, collectionVal)

	// Generar embeddings vectoriales reales usando text-embedding-004
	embedModel := b.Gemini.Client.EmbeddingModel("text-embedding-004")
	resEmbed, err := embedModel.EmbedContent(ctx, genai.Text(query))
	if err != nil {
		return "", fmt.Errorf("error al generar embeddings para búsqueda vectorial: %w", err)
	}

	queryVector := resEmbed.Embedding.Values
	log.Printf("Embedding generado correctamente (dim: %d). Inyectando en pipeline...", len(queryVector))

	// Parsear el pipeline JSON enviado por Gemini
	var pipelineObj []interface{}
	if err := json.Unmarshal([]byte(pipelineStr), &pipelineObj); err != nil {
		// Fallback por si Gemini envía el pipeline como objeto individual en vez de array
		var singleStage map[string]interface{}
		if err2 := json.Unmarshal([]byte(pipelineStr), &singleStage); err2 == nil {
			pipelineObj = []interface{}{singleStage}
		} else {
			return "", fmt.Errorf("error al deserializar pipeline JSON enviado por Gemini: %w", err)
		}
	}

	// Inyectar de manera recursiva el vector generado en el campo 'queryVector'
	injectQueryVector(pipelineObj, queryVector)

	// Ejecutar la agregación en el MCP real de MongoDB
	resultadoAgregacion, err := mcp.EjecutarAggregate(ctx, dbVal, collectionVal, pipelineObj)
	if err != nil {
		return "", fmt.Errorf("error en ejecución de agregación en MongoDB MCP: %w", err)
	}

	// Reenviar a Gemini para la síntesis final del reporte
	var parts []genai.Part
	parts = append(parts, genai.Text(query))
	parts = append(parts, resp.Candidates[0].Content.Parts...)
	parts = append(parts, genai.FunctionResponse{
		Name: "aggregate",
		Response: map[string]interface{}{
			"result": resultadoAgregacion,
		},
	})

	respFinal, err := model.GenerateContent(ctx, parts...)
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

// Función recursiva para buscar e inyectar el queryVector en la estructura de agregación
func injectQueryVector(v interface{}, queryVector []float32) {
	switch val := v.(type) {
	case map[string]interface{}:
		for k, item := range val {
			if k == "queryVector" {
				val[k] = queryVector
			} else {
				injectQueryVector(item, queryVector)
			}
		}
	case []interface{}:
		for _, item := range val {
			injectQueryVector(item, queryVector)
		}
	}
}
