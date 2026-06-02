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

// CronometradoraAgent orquesta la observabilidad y telemetría de rendimiento usando Elastic MCP y Gemini
type CronometradoraAgent struct {
	Gemini *llm.GeminiClient
}

// NewCronometradoraAgent inicializa un nuevo agente de La Cronometradora
func NewCronometradoraAgent(client *llm.GeminiClient) *CronometradoraAgent {
	return &CronometradoraAgent{
		Gemini: client,
	}
}

// AnalizarRendimiento analiza el rendimiento del código usando telemetría indexada en Elastic
func (c *CronometradoraAgent) AnalizarRendimiento(ctx context.Context, codigo string) (string, error) {
	systemInstruction := "Eres La Cronometradora. Analizas el rendimiento del código Go del estudiante. Usa tu herramienta analizar_telemetria_elastic para obtener las métricas indexadas. Basándote ÚNICAMENTE en esos datos, felicita al estudiante si es O(1) u O(n), o adviértele amablemente si es O(n^2) sugiriendo cómo optimizarlo."

	// 1. Caso de simulación o mock si el cliente no está inicializado
	if c.Gemini == nil || c.Gemini.Client == nil {
		log.Println("Cronometradora [MOCK]: Cliente Gemini no configurado. Simulando Tool Calling a analizar_telemetria_elastic...")

		// Obtener métricas del cliente MCP de Elastic
		metricasJSON, err := mcp.IndexarMetricasRendimiento(codigo)
		if err != nil {
			return "", fmt.Errorf("error al obtener métricas de telemetría: %w", err)
		}

		var telemetry mcp.ElasticTelemetry
		if err := json.Unmarshal([]byte(metricasJSON), &telemetry); err != nil {
			return "", fmt.Errorf("error decodificando telemetría de Elastic: %w", err)
		}

		var respuestaSimulada string
		if telemetry.Complexity == "O(n^2)" {
			respuestaSimulada = fmt.Sprintf("¡Reporte de Observabilidad de La Cronometradora! ⏱️\n\n**Métricas de Telemetría APM (Elastic):**\n- Complejidad: **O(n^2)**\n- Consumo de Memoria: **%s**\n- Tiempo de Ejecución: **%s**\n\n*Advertencia:* He detectado bucles anidados en tu solución. Esto causará una penalización de rendimiento en sistemas productivos de alta concurrencia. Sugiero reestructurar el algoritmo utilizando un mapa (hashtable) o variables auxiliares para reducir la complejidad a O(n).", telemetry.Memory, telemetry.ExecutionTime)
		} else {
			respuestaSimulada = fmt.Sprintf("¡Reporte de Observabilidad de La Cronometradora! ⏱️\n\n**Métricas de Telemetría APM (Elastic):**\n- Complejidad: **%s**\n- Consumo de Memoria: **%s**\n- Tiempo de Ejecución: **%s**\n\n*Felicitaciones:* Tu algoritmo tiene un perfil óptimo de rendimiento y bajo overhead de memoria. El runtime de Go ejecutará estas instrucciones de forma sumamente eficiente en producción.", telemetry.Complexity, telemetry.Memory, telemetry.ExecutionTime)
		}

		return respuestaSimulada, nil
	}

	// 2. Flujo real con la API de Gemini
	model := c.Gemini.Client.GenerativeModel("gemini-2.5-flash")
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(systemInstruction)},
	}

	// Declaración de la herramienta analizar_telemetria_elastic
	funcDecl := &genai.FunctionDeclaration{
		Name:        "analizar_telemetria_elastic",
		Description: "Indexa el código e interactúa con el servidor MCP de ElasticSearch para recuperar la telemetría APM de CPU, memoria y complejidad algorítmica.",
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"codigo": {
					Type:        genai.TypeString,
					Description: "El código fuente del estudiante para perfilar en el APM.",
				},
			},
			Required: []string{"codigo"},
		},
	}

	model.Tools = []*genai.Tool{
		{
			FunctionDeclarations: []*genai.FunctionDeclaration{funcDecl},
		},
	}

	// Enviar la consulta inicial
	resp, err := model.GenerateContent(ctx, genai.Text("Analiza el rendimiento del siguiente código:\n\n"+codigo))
	if err != nil {
		return "", fmt.Errorf("error en llamada inicial a Cronometradora: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return "No se ha recibido reporte de observabilidad del agente Cronometradora.", nil
	}

	part := resp.Candidates[0].Content.Parts[0]

	// Verificar si Gemini ha invocado la herramienta analizar_telemetria_elastic
	funcCall, ok := part.(genai.FunctionCall)
	if !ok {
		if textPart, isText := part.(genai.Text); isText {
			return string(textPart), nil
		}
		return "Respuesta vacía o formato inesperado de Gemini en Cronometradora.", nil
	}

	if funcCall.Name != "analizar_telemetria_elastic" {
		return "", fmt.Errorf("Gemini invocó una herramienta inesperada: %s", funcCall.Name)
	}

	// Extraer argumento
	codigoVal, exists := funcCall.Args["codigo"]
	if !exists {
		return "", fmt.Errorf("falta el parámetro requerido 'codigo' en analizar_telemetria_elastic")
	}
	codigoAAnalizar, ok := codigoVal.(string)
	if !ok {
		return "", fmt.Errorf("el parámetro 'codigo' debe ser un string")
	}

	log.Printf("Cronometradora [REAL]: Gemini solicitó telemetría de observabilidad en Elastic.")

	// Obtener métricas reales simuladas en el cliente MCP de Elastic
	metricasJSON, err := mcp.IndexarMetricasRendimiento(codigoAAnalizar)
	if err != nil {
		return "", fmt.Errorf("error al interactuar con Elastic MCP: %w", err)
	}

	// Segunda interacción: Devolver los resultados a Gemini para la síntesis de observabilidad
	var parts []genai.Part
	parts = append(parts, genai.Text("Analiza el rendimiento del siguiente código:\n\n"+codigo))
	parts = append(parts, resp.Candidates[0].Content.Parts...)
	parts = append(parts, genai.FunctionResponse{
		Name: "analizar_telemetria_elastic",
		Response: map[string]interface{}{
			"result": metricasJSON,
		},
	})

	respFinal, err := model.GenerateContent(ctx, parts...)
	if err != nil {
		return "", fmt.Errorf("error al sintetizar reporte de observabilidad con Gemini: %w", err)
	}

	if len(respFinal.Candidates) == 0 || respFinal.Candidates[0].Content == nil || len(respFinal.Candidates[0].Content.Parts) == 0 {
		return "No se pudo sintetizar el veredicto final de observabilidad.", nil
	}

	finalPart := respFinal.Candidates[0].Content.Parts[0]
	if textFinal, isText := finalPart.(genai.Text); isText {
		return string(textFinal), nil
	}

	return "No se pudo decodificar la respuesta final de La Cronometradora.", nil
}
