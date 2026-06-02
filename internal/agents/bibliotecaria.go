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

// BibliotecariaAgent orquesta la interacción con Gemini y la llamada a herramientas (Tool Calling)
type BibliotecariaAgent struct {
	Gemini *llm.GeminiClient
}

// NewBibliotecariaAgent inicializa un nuevo agente de La Bibliotecaria
func NewBibliotecariaAgent(client *llm.GeminiClient) *BibliotecariaAgent {
	return &BibliotecariaAgent{
		Gemini: client,
	}
}

// ConsultarDocumentacion ejecuta el flujo completo de RAG con Tool Calling
func (b *BibliotecariaAgent) ConsultarDocumentacion(ctx context.Context, query string) (string, error) {
	systemInstruction := "Eres La Bibliotecaria del ecosistema GOland. Tu única herramienta es buscar en la base de datos de MongoDB. Nunca inventes código. Responde solo basándote en el contexto devuelto por tu herramienta. Si no hay contexto relevante, indícalo educadamente."

	// 1. Caso de simulación o mock si el cliente no está inicializado
	if b.Gemini == nil || b.Gemini.Client == nil {
		log.Println("Bibliotecaria [MOCK]: Cliente Gemini no configurado. Simulando Tool Calling a search_mongodb_docs...")
		
		// Simular la llamada a la herramienta MCP local
		contextoSimulado, err := mcp.EjecutarBusquedaVectorial(query)
		if err != nil {
			return "", fmt.Errorf("error en búsqueda simulada: %w", err)
		}

		// Simulación de la síntesis final del agente basado en el contexto
		var docs []mcp.DocumentoDoc
		if err := json.Unmarshal([]byte(contextoSimulado), &docs); err != nil {
			return "", fmt.Errorf("error decodificando contexto simulado: %w", err)
		}

		var respuestaSimulada string
		if len(docs) > 0 {
			respuestaSimulada = fmt.Sprintf("¡Hola! Soy La Bibliotecaria de GOland. He consultado el servidor MCP de MongoDB Atlas y encontré esto sobre tu consulta:\n\n**%s**:\n%s\n\n*Fuente oficial:* [Ir a la documentación](%s) (Confianza: %.2f)", 
				docs[0].Titulo, docs[0].Contenido, docs[0].Fuente, docs[0].Score)
		} else {
			respuestaSimulada = "Lo siento, no he podido encontrar información relevante sobre tu consulta en la documentación indexada en MongoDB."
		}

		return respuestaSimulada, nil
	}

	// 2. Flujo real con la API de Gemini
	model := b.Gemini.Client.GenerativeModel("gemini-2.5-flash")
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(systemInstruction)},
	}

	// Declaración de la herramienta
	funcDecl := &genai.FunctionDeclaration{
		Name:        "search_mongodb_docs",
		Description: "Busca en la base de datos de documentación de MongoDB Atlas Vector Search.",
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"query": {
					Type:        genai.TypeString,
					Description: "El término de búsqueda o pregunta de documentación sobre Go.",
				},
			},
			Required: []string{"query"},
		},
	}

	model.Tools = []*genai.Tool{
		{
			FunctionDeclarations: []*genai.FunctionDeclaration{funcDecl},
		},
	}

	// Primera interacción: Enviar la consulta inicial a Gemini
	resp, err := model.GenerateContent(ctx, genai.Text(query))
	if err != nil {
		return "", fmt.Errorf("error en llamada inicial a Gemini: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return "No se ha recibido respuesta del agente.", nil
	}

	part := resp.Candidates[0].Content.Parts[0]

	// Verificar si Gemini ha decidido invocar la herramienta search_mongodb_docs
	funcCall, ok := part.(genai.FunctionCall)
	if !ok {
		// Si Gemini respondió directamente sin usar herramientas (lo cual viola la System Instruction pero es posible)
		if textPart, isText := part.(genai.Text); isText {
			return string(textPart), nil
		}
		return "Respuesta vacía o formato inesperado de Gemini.", nil
	}

	if funcCall.Name != "search_mongodb_docs" {
		return "", fmt.Errorf("Gemini invocó una herramienta inesperada: %s", funcCall.Name)
	}

	// Extraer argumento de la llamada de función
	searchQueryVal, exists := funcCall.Args["query"]
	if !exists {
		return "", fmt.Errorf("falta el parámetro requerido 'query' en la invocación")
	}
	searchQuery, ok := searchQueryVal.(string)
	if !ok {
		return "", fmt.Errorf("el parámetro 'query' debe ser un string")
	}

	log.Printf("Bibliotecaria [REAL]: Gemini solicitó búsqueda en MongoDB con query: '%s'", searchQuery)

	// Ejecutar la búsqueda vectorial en el cliente MCP de MongoDB
	resultadoBusqueda, err := mcp.EjecutarBusquedaVectorial(searchQuery)
	if err != nil {
		return "", fmt.Errorf("error al ejecutar búsqueda vectorial en MongoDB MCP: %w", err)
	}

	// Segunda interacción: Devolver los resultados obtenidos a Gemini para que sintetice la respuesta
	var parts []genai.Part
	parts = append(parts, genai.Text(query))
	parts = append(parts, resp.Candidates[0].Content.Parts...)
	parts = append(parts, genai.FunctionResponse{
		Name: "search_mongodb_docs",
		Response: map[string]interface{}{
			"result": resultadoBusqueda,
		},
	})

	respFinal, err := model.GenerateContent(ctx, parts...)
	if err != nil {
		return "", fmt.Errorf("error en respuesta final de Gemini tras proveer contexto: %w", err)
	}

	if len(respFinal.Candidates) == 0 || respFinal.Candidates[0].Content == nil || len(respFinal.Candidates[0].Content.Parts) == 0 {
		return "No se pudo sintetizar el resultado final.", nil
	}

	finalPart := respFinal.Candidates[0].Content.Parts[0]
	if textFinal, isText := finalPart.(genai.Text); isText {
		return string(textFinal), nil
	}

	return "No se pudo decodificar la respuesta final de Gemini.", nil
}
