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

// ConstructorAgent orquesta la validación de código con Gemini y pipelines de CI/CD mediante GitLab MCP
type ConstructorAgent struct {
	Gemini *llm.GeminiClient
}

// NewConstructorAgent inicializa un nuevo agente de El Constructor
func NewConstructorAgent(client *llm.GeminiClient) *ConstructorAgent {
	return &ConstructorAgent{
		Gemini: client,
	}
}

// ValidarCodigo realiza la validación ejecutando el pipeline simulado mediante GitLab MCP
func (c *ConstructorAgent) ValidarCodigo(ctx context.Context, codigo string) (string, error) {
	systemInstruction := "Eres El Constructor. Tu trabajo es validar el código Go del estudiante. NO evalúes el código a ciegas. USA SIEMPRE tu herramienta trigger_gitlab_ci para compilarlo. Basándote ÚNICAMENTE en el log de compilación que te devuelva la herramienta, aprueba al estudiante o explícale el error de compilación."

	// 1. Caso de simulación o mock si el cliente no está inicializado
	if c.Gemini == nil || c.Gemini.Client == nil {
		log.Println("Constructor [MOCK]: Cliente Gemini no configurado. Simulando Tool Calling a trigger_gitlab_ci...")

		// Simular ejecución del pipeline de GitLab
		logPipeline, err := mcp.EjecutarPipelineGo(codigo)
		if err != nil {
			return "", fmt.Errorf("error en pipeline simulado: %w", err)
		}

		var res mcp.PipelineResult
		if err := json.Unmarshal([]byte(logPipeline), &res); err != nil {
			return "", fmt.Errorf("error decodificando log de pipeline simulado: %w", err)
		}

		var respuestaSimulada string
		if res.Status == "success" {
			respuestaSimulada = fmt.Sprintf("¡Veredicto de El Constructor! 🛠️\n\nEl pipeline de CI/CD de GitLab se ha ejecutado exitosamente:\n\n`%s`\n\n¡Felicidades! Tu código compila sin errores y ha pasado la batería de pruebas en el contenedor virtual.", res.Log)
		} else {
			respuestaSimulada = fmt.Sprintf("¡Alerta de Compilación de El Constructor! ❌\n\nEl pipeline de GitLab falló con el siguiente registro de salida:\n\n```\n%s\n```\n\nPor favor, revisa la sintaxis indicada en el log y corrige los errores en tu editor.", res.Log)
		}

		return respuestaSimulada, nil
	}

	// 2. Flujo real con la API de Gemini
	model := c.Gemini.Client.GenerativeModel("gemini-2.5-flash")
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(systemInstruction)},
	}

	// Declaración de la herramienta trigger_gitlab_ci
	funcDecl := &genai.FunctionDeclaration{
		Name:        "trigger_gitlab_ci",
		Description: "Envía el código Go del estudiante a un pipeline de GitLab CI/CD para compilarlo y probarlo.",
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"codigo_estudiante": {
					Type:        genai.TypeString,
					Description: "El código fuente completo de Go que se compilará.",
				},
			},
			Required: []string{"codigo_estudiante"},
		},
	}

	model.Tools = []*genai.Tool{
		{
			FunctionDeclarations: []*genai.FunctionDeclaration{funcDecl},
		},
	}

	// Enviar la consulta inicial que solicita la validación del código
	resp, err := model.GenerateContent(ctx, genai.Text("Por favor, valida este código:\n\n"+codigo))
	if err != nil {
		return "", fmt.Errorf("error en llamada inicial de compilación a Gemini: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return "No se ha recibido veredicto del agente Constructor.", nil
	}

	part := resp.Candidates[0].Content.Parts[0]

	// Verificar si Gemini ha invocado la herramienta trigger_gitlab_ci
	funcCall, ok := part.(genai.FunctionCall)
	if !ok {
		// Fallback si no decide llamar a la herramienta
		if textPart, isText := part.(genai.Text); isText {
			return string(textPart), nil
		}
		return "Respuesta vacía o formato inesperado de Gemini en Constructor.", nil
	}

	if funcCall.Name != "trigger_gitlab_ci" {
		return "", fmt.Errorf("Gemini invocó una herramienta de compilación inesperada: %s", funcCall.Name)
	}

	// Extraer argumento de la llamada de función
	codigoVal, exists := funcCall.Args["codigo_estudiante"]
	if !exists {
		return "", fmt.Errorf("falta el parámetro requerido 'codigo_estudiante' en la llamada a trigger_gitlab_ci")
	}
	codigoEstudiante, ok := codigoVal.(string)
	if !ok {
		return "", fmt.Errorf("el parámetro 'codigo_estudiante' debe ser un string")
	}

	log.Printf("Constructor [REAL]: Gemini solicitó ejecución de pipeline GitLab para código del alumno.")

	// Ejecutar compilación remota simulada en el cliente MCP de GitLab
	logCompilacion, err := mcp.EjecutarPipelineGo(codigoEstudiante)
	if err != nil {
		return "", fmt.Errorf("error ejecutando pipeline GitLab CI: %w", err)
	}

	// Devolver el log de compilación a Gemini para la síntesis final
	var parts []genai.Part
	parts = append(parts, genai.Text("Por favor, valida este código:\n\n"+codigo))
	parts = append(parts, resp.Candidates[0].Content.Parts...)
	parts = append(parts, genai.FunctionResponse{
		Name: "trigger_gitlab_ci",
		Response: map[string]interface{}{
			"result": logCompilacion,
		},
	})

	respFinal, err := model.GenerateContent(ctx, parts...)
	if err != nil {
		return "", fmt.Errorf("error al sintetizar reporte de compilación con Gemini: %w", err)
	}

	if len(respFinal.Candidates) == 0 || respFinal.Candidates[0].Content == nil || len(respFinal.Candidates[0].Content.Parts) == 0 {
		return "No se pudo sintetizar el reporte final del compilador.", nil
	}

	finalPart := respFinal.Candidates[0].Content.Parts[0]
	if textFinal, isText := finalPart.(genai.Text); isText {
		return string(textFinal), nil
	}

	return "No se pudo decodificar la respuesta final de El Constructor.", nil
}
