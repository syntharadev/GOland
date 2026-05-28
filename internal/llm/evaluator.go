package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/generative-ai-go/genai"
)

// Evaluacion define la estructura estricta del veredicto del enjambre
type Evaluacion struct {
	Aprobado            bool   `json:"aprobado"`
	Feedback            string `json:"feedback"`
	GOnionInterviniente string `json:"gonion_interviniente"` // Qué agente da la respuesta
}

// EvaluateAndProgress somete el código del usuario al juicio del Consejo
func (g *GeminiClient) EvaluateAndProgress(ctx context.Context, nick, dominio, objetivo string, nivelActual int, codigoUsuario string) (*Evaluacion, error) {
	// Fallback de seguridad si el cliente no fue inicializado con API Key real
	if g.Client == nil || fmt.Sprintf("%v", g.Client) == "&{}" {
		log.Println("Mock: Retornando evaluación quemada")
		return &Evaluacion{
			Aprobado:            true,
			Feedback:            "¡Excelente trabajo! Has implementado correctamente el código.",
			GOnionInterviniente: "Astro-Gopher",
		}, nil
	}

	model := g.Client.GenerativeModel("gemini-2.5-flash")
	model.ResponseMIMEType = "application/json"

	systemPrompt := `Eres un Evaluador Senior de código Go dentro del sistema 'GOland'.
Tu tarea es analizar el código escrito por el usuario y determinar si cumple con el objetivo del nivel.
Debes elegir a uno de los agentes del Consejo (ej. 'Agente Sintaxis' o 'Gopher Seguridad') para que dé el feedback.
Si el código tiene errores de compilación, lógica o no cumple el objetivo, 'aprobado' debe ser false.
Si es correcto y eficiente, 'aprobado' debe ser true.
Devuelve estrictamente un JSON válido con esta estructura:
{
  "aprobado": true/false,
  "feedback": "Tu explicación narrativa y técnica aquí...",
  "gonion_interviniente": "Nombre del Agente"
}`

	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(systemPrompt)},
	}

	promptTexto := fmt.Sprintf("Piloto: %s\nObjetivo del Proyecto: %s\nNivel Actual: %d\n\nCódigo del Usuario:\n```go\n%s\n```", nick, objetivo, nivelActual, codigoUsuario)
	resp, err := model.GenerateContent(ctx, genai.Text(promptTexto))
	if err != nil {
		return nil, fmt.Errorf("error al evaluar código: %w", err)
	}

	jsonText, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return nil, fmt.Errorf("formato de respuesta inesperado")
	}

	var eval Evaluacion
	if err := json.Unmarshal([]byte(jsonText), &eval); err != nil {
		log.Printf("Error decodificando evaluación JSON: %s", jsonText)
		return nil, fmt.Errorf("error de unmarshal: %w", err)
	}

	return &eval, nil
}
