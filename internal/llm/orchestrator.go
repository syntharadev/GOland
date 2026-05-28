package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/generative-ai-go/genai"
)

type GOnion struct {
	Nombre             string `json:"nombre"`
	Familia            string `json:"familia"`
	Nivel              int    `json:"nivel"`
	Personalidad       string `json:"personalidad"`
	AparienciaSugerida string `json:"apariencia_sugerida"`
}

type RetoInicial struct {
	MensajeGOnion string `json:"mensaje_gonion"`
	CodigoBase    string `json:"codigo_base"` // Plantilla inicial para el textarea
	Instrucciones string `json:"instrucciones"`
}

type SwarmConfiguration struct {
	Tema           string      `json:"tema"`
	NivelesTotal   int         `json:"niveles_total"`
	ConsejoGOnions []GOnion    `json:"consejo_gonions"`
	HiloConductor  string      `json:"hilo_conductor"`
	RetoActual     RetoInicial `json:"reto_actual"` // NUEVO: El desafío inicial integrado
}

func (g *GeminiClient) GenerateSwarmWorld(ctx context.Context, nick, nivelExperiencia, dominio, objetivo string, nivelActual int) (*SwarmConfiguration, error) {
	// Fallback de seguridad si el cliente no fue inicializado con API Key real
	if g.Client == nil || fmt.Sprintf("%v", g.Client) == "&{}" {
		log.Println("Mock: Retornando datos quemados porque no hay cliente Gemini real configurado")
		return &SwarmConfiguration{
			Tema:          "Misión de prueba",
			NivelesTotal:  5,
			HiloConductor: "Iniciando secuencia de arranque para " + nick,
			ConsejoGOnions: []GOnion{
				{Nombre: "Astro-Gopher", Familia: "Backend", Nivel: 1, Personalidad: "Estricto", AparienciaSugerida: "Gopher Astronauta"},
				{Nombre: "GOnion Sec", Familia: "Seguridad", Nivel: 1, Personalidad: "Paranoico", AparienciaSugerida: "Gopher con escudo"},
				{Nombre: "Data GOnion", Familia: "Data", Nivel: 1, Personalidad: "Analítico", AparienciaSugerida: "Gopher con gafas"},
			},
			RetoActual: RetoInicial{
				MensajeGOnion: fmt.Sprintf("Piloto %s, para iniciar la telemetría, necesitamos crear nuestro primer paquete en Go.", nick),
				CodigoBase:    "package main\n\nimport \"fmt\"\n\nfunc main() {\n\t// Tu código aquí\n}",
				Instrucciones: "Escribe una función que imprima 'Sistemas Online'",
			},
		}, nil
	}

	model := g.Client.GenerativeModel("gemini-2.5-flash")
	model.ResponseMIMEType = "application/json"
	
	systemPrompt := fmt.Sprintf(`Eres el Orquestador de 'GOland'. 
Piloto: '%s' | Experiencia Previa: '%s'. 
Dominio: %s | Objetivo: %s.
ESTADO ACTUAL DE LA SESIÓN: Nivel %d.

TAREAS:
1. Crea/Reconstruye el Consejo de 3 GOnions temáticos.
2. Genera el 'reto_actual' para el NIVEL %d explícitamente.
   - Haz que un GOnion reciba al piloto en el 'mensaje_gonion' reconociendo su nivel actual de progreso.
Devuelve estrictamente el JSON con la estructura solicitada.`, nick, nivelExperiencia, dominio, objetivo, nivelActual, nivelActual)
	
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(systemPrompt)},
	}

	resp, err := model.GenerateContent(ctx, genai.Text("Inicia la simulación y el primer reto."))
	if err != nil {
		return nil, fmt.Errorf("error generando mundo: %w", err)
	}

	jsonText, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return nil, fmt.Errorf("formato de respuesta inesperado")
	}

	var config SwarmConfiguration
	if err := json.Unmarshal([]byte(jsonText), &config); err != nil {
		log.Printf("Error decodificando JSON del Orquestador: %s", jsonText)
		return nil, fmt.Errorf("error de unmarshal: %w", err)
	}

	return &config, nil
}
