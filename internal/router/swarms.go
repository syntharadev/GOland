package router

import (
	"context"
	"fmt"
	"log"
	"strings"

	"gemini-go-platform/internal/llm"
	"gemini-go-platform/internal/mcp"

	"github.com/google/generative-ai-go/genai"
)

type MensajeSwarm struct {
	Nombre string `json:"nombre"`
	Texto  string `json:"texto"`
	Avatar string `json:"avatar"`
}

// EjecutarSwarm despacha la consulta y el código al escuadrón correspondiente usando agentes Gemini dinámicos
func EjecutarSwarm(ctx context.Context, gemini *llm.GeminiClient, escuadron int, input string, codigo string, nick string) ([]MensajeSwarm, error) {
	var respuestas []MensajeSwarm

	switch escuadron {
	case 1:
		// Swarm 1 (Teoría/Dudas) -> La Bibliotecaria y El Hacker
		resBib, err := ejecutarAgenteGemini(ctx, gemini, "La Bibliotecaria", input, codigo, nick)
		if err != nil {
			resBib = fmt.Sprintf("Error en La Bibliotecaria: %v", err)
		}
		respuestas = append(respuestas, MensajeSwarm{
			Nombre: "La Bibliotecaria",
			Texto:  resBib,
			Avatar: "/static/img/Bibliotecaria.png",
		})

		resHacker, err := ejecutarAgenteGemini(ctx, gemini, "El Hacker", input, codigo, nick)
		if err != nil {
			resHacker = fmt.Sprintf("Error en El Hacker: %v", err)
		}
		respuestas = append(respuestas, MensajeSwarm{
			Nombre: "El Hacker",
			Texto:  resHacker,
			Avatar: "/static/img/Hacker.png",
		})

	case 2:
		// Swarm 2 (Arquitectura/Estructura) -> La Ingeniera y El Constructor
		resIng, err := ejecutarAgenteGemini(ctx, gemini, "La Ingeniera", input, codigo, nick)
		if err != nil {
			resIng = fmt.Sprintf("Error en La Ingeniera: %v", err)
		}
		respuestas = append(respuestas, MensajeSwarm{
			Nombre: "La Ingeniera",
			Texto:  resIng,
			Avatar: "/static/img/Ingeniera.png",
		})

		resConst, err := ejecutarAgenteGemini(ctx, gemini, "El Constructor", input, codigo, nick)
		if err != nil {
			resConst = fmt.Sprintf("Error en El Constructor: %v", err)
		}
		respuestas = append(respuestas, MensajeSwarm{
			Nombre: "El Constructor",
			Texto:  resConst,
			Avatar: "/static/img/Constructor.png",
		})

	case 3:
		// Swarm 3 (Sintaxis/Logs) -> El Senior y El Mensajero
		resSenior, err := ejecutarAgenteGemini(ctx, gemini, "El Senior", input, codigo, nick)
		if err != nil {
			resSenior = fmt.Sprintf("Error en El Senior: %v", err)
		}
		respuestas = append(respuestas, MensajeSwarm{
			Nombre: "El Senior",
			Texto:  resSenior,
			Avatar: "/static/img/Senior.png",
		})

		resMensajero, err := ejecutarAgenteGemini(ctx, gemini, "El Mensajero", input, codigo, nick)
		if err != nil {
			resMensajero = fmt.Sprintf("Error en El Mensajero: %v", err)
		}
		respuestas = append(respuestas, MensajeSwarm{
			Nombre: "El Mensajero",
			Texto:  resMensajero,
			Avatar: "/static/img/Mensajero.png",
		})

	case 4:
		// Swarm 4 (Optimización/Tests) -> La Guardiana y La Cronometradora
		resGuardiana, err := ejecutarAgenteGemini(ctx, gemini, "La Guardiana", input, codigo, nick)
		if err != nil {
			resGuardiana = fmt.Sprintf("Error en La Guardiana: %v", err)
		}
		respuestas = append(respuestas, MensajeSwarm{
			Nombre: "La Guardiana",
			Texto:  resGuardiana,
			Avatar: "/static/img/Guardiana.png",
		})

		resChrono, err := ejecutarAgenteGemini(ctx, gemini, "La Cronometradora", input, codigo, nick)
		if err != nil {
			resChrono = fmt.Sprintf("Error en La Cronometradora: %v", err)
		}
		respuestas = append(respuestas, MensajeSwarm{
			Nombre: "La Cronometradora",
			Texto:  resChrono,
			Avatar: "/static/img/Cronometradora.png",
		})
	}

	return respuestas, nil
}

// ejecutarAgenteGemini encapsula la llamada real/simulada de Gemini para cualquier agente aplicando System Instructions, Temperature y Chat History.
func ejecutarAgenteGemini(ctx context.Context, gemini *llm.GeminiClient, agente string, input string, codigo string, nick string) (string, error) {
	var systemPrompt string
	var temp float32 = 0.5

	switch agente {
	case "La Bibliotecaria":
		systemPrompt = "Eres La Bibliotecaria del ecosistema GOland. Tu objetivo es ayudar a los estudiantes resolviendo sus dudas de programación en Go. Responde de forma clara y didáctica basándote estrictamente en el contexto proporcionado. Si el contexto está vacío o no es relevante, indícalo de forma educada y responde según tus conocimientos generales de Go, aclarando que no se encontró documentación oficial al respecto en el RAG."
		temp = 0.15
	case "La Guardiana":
		systemPrompt = "Eres La Guardiana. Tu tono es implacable, enfocado exclusivamente en encontrar casos extremos (edge cases), fugas de memoria y errores lógicos en el código del estudiante. NO des código de solución bajo ninguna circunstancia. Tu misión es empujar al estudiante a pensar de forma crítica en los límites de su código."
		temp = 0.1
	case "El Mensajero":
		systemPrompt = "Eres El Mensajero. Tu personalidad es veloz, concisa y técnica. Analizas los logs de consola y notificaciones de sistema. Reportas diagnósticos y eventos de red con estilo de despacho y telemetría clara."
		temp = 0.1
	case "La Ingeniera":
		systemPrompt = "Eres La Ingeniera. Tu personalidad es sumamente analítica, estructurada y orientada al diseño limpio. Analizas el rendimiento, concurrencia (goroutines), el diseño de interfaces y la optimización de memoria. Explica la arquitectura recomendada usando buenas prácticas de desacoplamiento en Go."
		temp = 0.2
	case "El Profesor":
		systemPrompt = "Eres El Profesor, el nodo central de GOland. Tu tono es sumamente socrático, didáctico y paciente. Guías al estudiante paso a paso hacia la respuesta sin darle la solución directamente. Hazle preguntas que lo inviten a razonar."
		temp = 0.5
	case "El Constructor":
		systemPrompt = "Eres El Constructor. Tu personalidad es pragmática, orientada a la compilación y herramientas. Tu trabajo es validar el código Go del estudiante. NO evalúes el código a ciegas. USA tu herramienta trigger_gitlab_ci para compilarlo. Basándote ÚNICAMENTE en el log de compilación que te devuelva la herramienta, aprueba al estudiante o explícale el error de compilación."
		temp = 0.5
	case "El Senior":
		systemPrompt = "Eres El Senior. Tu tono es experimentado, un poco rudo pero sabio. Analizas las convenciones de nomenclatura (Effective Go), legibilidad y buenas prácticas generales de Go. Exige control estricto de errores y simplicidad extrema."
		temp = 0.2
	case "La Cronometradora":
		systemPrompt = "Eres La Cronometradora. Tu personalidad es precisa, obsesionada con el tiempo de ejecución y la notación Big O. Analizas el rendimiento del código Go usando telemetría APM (CPU y memoria). Explica la complejidad algorítmica e indica si hay bucles anidados u oportunidades de optimización con estructuras de datos eficientes."
		temp = 0.1
	case "El Hacker":
		systemPrompt = "Eres El Hacker. Tu personalidad es irreverente, astuta e ingeniosa. Enseñas atajos prácticos, trucos no documentados, one-liners y optimizaciones ingeniosas en Go. Tu enfoque es creativo y out-of-the-box."
		temp = 1.0
	default:
		systemPrompt = "Eres un tutor de Go experto y amigable."
		temp = 0.5
	}

	// 1. Simulación si el cliente es nulo
	if gemini == nil || gemini.Client == nil {
		log.Printf("[MOCK] Agente %s: Cliente Gemini no inicializado.", agente)
		switch agente {
		case "La Bibliotecaria":
			return fmt.Sprintf("¡Hola! Soy La Bibliotecaria de GOland (Modo Simulación). He consultado el RAG sobre tu duda: '%s'. ¡Todo funciona!", input), nil
		case "La Guardiana":
			return "¡La Guardiana al habla! He auditado tu código conceptualmente. Ten cuidado con los edge cases y el manejo de nil pointers. 🛡️", nil
		case "El Mensajero":
			return "¡Mensaje del sistema! Todo el tráfico de red y los logs fluyen estables.", nil
		case "La Ingeniera":
			return "Te sugiero separar la inicialización de tus servicios de Go en subpaquetes modulares.", nil
		case "El Profesor":
			return "¿Has intentado declarar esa variable usando var o el operador :=? Cuéntame tu razonamiento.", nil
		case "El Constructor":
			return "Compilación exitosa en el pipeline de GitLab (Simulado).", nil
		case "El Senior":
			return "Recuerda controlar siempre los errores en Go ('if err != nil'). Es la regla de oro.", nil
		case "La Cronometradora":
			return "Complejidad detectada: O(n). Excelente consumo de CPU y memoria.", nil
		case "El Hacker":
			return "Tip rápido: Puedes usar channels con buffer para evitar bloqueos innecesarios en go-routines.", nil
		default:
			return "Modo Simulación Activo.", nil
		}
	}

	// 2. Obtener o crear ChatSession de Gemini para este par usuario-agente (Memoria de Sesión)
	sessionKey := fmt.Sprintf("%s_%s", nick, agente)
	var cs *genai.ChatSession

	if val, ok := gemini.ChatSessions.Load(sessionKey); ok {
		cs = val.(*genai.ChatSession)
	} else {
		model := gemini.Client.GenerativeModel("gemini-2.5-flash")
		model.SystemInstruction = &genai.Content{
			Parts: []genai.Part{genai.Text(systemPrompt)},
		}
		model.Temperature = &temp

		// Declarar herramientas solo para los agentes que las requieran
		if agente == "El Constructor" {
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
		} else if agente == "La Cronometradora" {
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
		}

		cs = model.StartChat()
		gemini.ChatSessions.Store(sessionKey, cs)
	}

	// 3. Inyección RAG para La Bibliotecaria
	var prompt string
	if agente == "La Bibliotecaria" {
		words := strings.Fields(input)
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
			regexPattern = "(?i)" + input
		}

		log.Printf("Bibliotecaria RAG: Patrón de búsqueda: '%s'", regexPattern)
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

		resultadoAgregacion, err := mcp.EjecutarAggregate(ctx, "goland_db", "go_docs", pipeline)
		if err != nil {
			log.Printf("Error RAG: %v", err)
			resultadoAgregacion = "No se pudo recuperar contexto de la base de datos."
		}

		if codigo != "" {
			prompt = fmt.Sprintf("Contexto RAG de MongoDB:\n%s\n\nMensaje del usuario: %s\nContexto del código actual:\n%s", resultadoAgregacion, input, codigo)
		} else {
			prompt = fmt.Sprintf("Contexto RAG de MongoDB:\n%s\n\nMensaje del usuario: %s", resultadoAgregacion, input)
		}
	} else {
		// Inyección de Contexto Oculta estándar
		if codigo != "" {
			prompt = fmt.Sprintf("Mensaje del usuario: %s\nContexto del código actual:\n%s", input, codigo)
		} else {
			prompt = fmt.Sprintf("Mensaje del usuario: %s", input)
		}
	}

	// 4. Enviar mensaje en la sesión de chat
	resp, err := cs.SendMessage(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return fmt.Sprintf("No se recibió veredicto de %s.", agente), nil
	}

	part := resp.Candidates[0].Content.Parts[0]

	// 5. Manejar Tool Calling si es necesario
	if funcCall, ok := part.(genai.FunctionCall); ok {
		if funcCall.Name == "trigger_gitlab_ci" {
			codigoVal := funcCall.Args["codigo_estudiante"]
			codigoEstudiante, _ := codigoVal.(string)
			logCompilacion, err := mcp.EjecutarPipelineGo(codigoEstudiante)
			if err != nil {
				return "", err
			}

			respFinal, err := cs.SendMessage(ctx, genai.FunctionResponse{
				Name: "trigger_gitlab_ci",
				Response: map[string]interface{}{
					"result": logCompilacion,
				},
			})
			if err != nil {
				return "", err
			}

			if len(respFinal.Candidates) > 0 && respFinal.Candidates[0].Content != nil && len(respFinal.Candidates[0].Content.Parts) > 0 {
				part = respFinal.Candidates[0].Content.Parts[0]
			}
		} else if funcCall.Name == "analizar_telemetria_elastic" {
			codigoVal := funcCall.Args["codigo"]
			codigoAAnalizar, _ := codigoVal.(string)
			metricasJSON, err := mcp.IndexarMetricasRendimiento(codigoAAnalizar)
			if err != nil {
				return "", err
			}

			respFinal, err := cs.SendMessage(ctx, genai.FunctionResponse{
				Name: "analizar_telemetria_elastic",
				Response: map[string]interface{}{
					"result": metricasJSON,
				},
			})
			if err != nil {
				return "", err
			}

			if len(respFinal.Candidates) > 0 && respFinal.Candidates[0].Content != nil && len(respFinal.Candidates[0].Content.Parts) > 0 {
				part = respFinal.Candidates[0].Content.Parts[0]
			}
		}
	}

	if textPart, isText := part.(genai.Text); isText {
		return string(textPart), nil
	}

	return "Formato de respuesta inesperado.", nil
}
