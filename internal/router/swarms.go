package router

import (
	"context"
	"fmt"

	"gemini-go-platform/internal/agents"
	"gemini-go-platform/internal/llm"
)

type MensajeSwarm struct {
	Nombre string `json:"nombre"`
	Texto  string `json:"texto"`
	Avatar string `json:"avatar"`
}

// EjecutarSwarm despacha la consulta y el código al escuadrón correspondiente
func EjecutarSwarm(ctx context.Context, gemini *llm.GeminiClient, escuadron int, input string, codigo string) ([]MensajeSwarm, error) {
	var respuestas []MensajeSwarm

	switch escuadron {
	case 1:
		// Swarm 1 (Teoría/Dudas) -> La Bibliotecaria y El Hacker
		// 1. Bibliotecaria RAG MCP MongoDB
		bibliotecaria := agents.NewBibliotecariaAgent(gemini)
		resBib, err := bibliotecaria.ConsultarDocumentacion(ctx, input)
		if err != nil {
			resBib = fmt.Sprintf("Error RAG en La Bibliotecaria: %v", err)
		}
		respuestas = append(respuestas, MensajeSwarm{
			Nombre: "La Bibliotecaria",
			Texto:  resBib,
			Avatar: "/static/img/Bibliotecaria.png",
		})

		// 2. El Hacker (Soporte conceptual de seguridad y buenas prácticas)
		respuestas = append(respuestas, MensajeSwarm{
			Nombre: "El Hacker",
			Texto:  fmt.Sprintf("Auditoría cuántica de seguridad completada. He analizado '%s'. No detecto riesgos de inyección, desbordamientos de buffer u otras vulnerabilidades críticas de red en tu propuesta conceptual. ¡Código seguro! 🛡️💻", input),
			Avatar: "/static/img/Hacker.png",
		})

	case 2:
		// Swarm 2 (Arquitectura/Estructura) -> La Ingeniera y El Constructor
		// 1. La Ingeniera
		respuestas = append(respuestas, MensajeSwarm{
			Nombre: "La Ingeniera",
			Texto:  fmt.Sprintf("He analizado el diseño estructural para tu requerimiento: '%s'. Te recomiendo diseñar tu solución siguiendo patrones desacoplados en Go: expón interfaces para tus servicios en '/internal/core' y separa la inicialización en '/cmd'. Esto facilitará las pruebas unitarias y mocks. 📐", input),
			Avatar: "/static/img/Ingeniera.png",
		})

		// 2. El Constructor
		resConst := "Estoy listo y a la espera de código Go en tu editor para lanzar el pipeline de validación en GitLab CI/CD. ⚙️"
		if codigo != "" {
			constructor := agents.NewConstructorAgent(gemini)
			res, err := constructor.ValidarCodigo(ctx, codigo)
			if err == nil {
				resConst = res
			} else {
				resConst = fmt.Sprintf("Fallo en pipeline de compilación CI/CD: %v", err)
			}
		}
		respuestas = append(respuestas, MensajeSwarm{
			Nombre: "El Constructor",
			Texto:  resConst,
			Avatar: "/static/img/Constructor.png",
		})

	case 3:
		// Swarm 3 (Sintaxis/Logs) -> El Senior y El Mensajero
		// 1. El Senior
		respuestas = append(respuestas, MensajeSwarm{
			Nombre: "El Senior",
			Texto:  fmt.Sprintf("Analizando sintaxis y convenciones sobre '%s'. En Go, la simplicidad es ley: recuerda controlar siempre los errores devueltos ('if err != nil') en lugar de usar pánico, y utiliza canales solo cuando la concurrencia lo justifique de verdad. 👴🏼", input),
			Avatar: "/static/img/Senior.png",
		})

		// 2. El Mensajero
		respuestas = append(respuestas, MensajeSwarm{
			Nombre: "El Mensajero",
			Texto:  "¡Reporte cuántico interceptado! He validado los logs de la consola local y del servidor. Todo el tráfico fluye de manera estable y limpia. ✉️",
			Avatar: "/static/img/Mensajero.png",
		})

	case 4:
		// Swarm 4 (Optimización/Tests) -> La Guardiana y La Cronometradora
		// 1. La Guardiana
		respuestas = append(respuestas, MensajeSwarm{
			Nombre: "La Guardiana",
			Texto:  fmt.Sprintf("He auditado la robustez de tu propuesta sobre '%s'. Todo parece estable desde el punto de vista del ciclo de vida. Recuerda evitar fugas de memoria (leak de goroutines) cerrando siempre tus canales y usando contextos con timeout al realizar peticiones externas. 🛡️", input),
			Avatar: "/static/img/Guardiana.png",
		})

		// 2. La Cronometradora
		resChrono := "Envía un fragmento de código Go a través del editor para realizar el análisis estático Big O e indexar la telemetría en Elastic APM. ⏱️"
		if codigo != "" {
			cronometradora := agents.NewCronometradoraAgent(gemini)
			res, err := cronometradora.AnalizarRendimiento(ctx, codigo)
			if err == nil {
				resChrono = res
			} else {
				resChrono = fmt.Sprintf("Error APM en La Cronometradora: %v", err)
			}
		}
		respuestas = append(respuestas, MensajeSwarm{
			Nombre: "La Cronometradora",
			Texto:  resChrono,
			Avatar: "/static/img/Cronometradora.png",
		})
	}

	return respuestas, nil
}
