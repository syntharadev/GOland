package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"gemini-go-platform/internal/database"
	"gemini-go-platform/internal/llm"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Temporal para desarrollo
	},
}

// MensajeEntrante define lo que esperamos recibir del Frontend
type MensajeEntrante struct {
	Tipo           string `json:"tipo"`            // ej: "INIT_WORLD"
	Nick           string `json:"nick"`            // NUEVO
	NivelPrevio    string `json:"nivel_previo"`    // NUEVO
	Dominio        string `json:"dominio"`         // ej: "Medicina"
	Objetivo       string `json:"objetivo"`        // ej: "Analizar datos genómicos"
	Codigo         string `json:"codigo"`          // ej: "func main() {}"
	ObjetivoActual string `json:"objetivo_actual"` // ej: "Crear una variable"
	NivelActual    int    `json:"nivel_actual"`    // NUEVO
}

// SwarmConnectionHandler ahora recibe el cliente Gemini y DB como dependencia
func SwarmConnectionHandler(w http.ResponseWriter, r *http.Request, gemini *llm.GeminiClient, database *database.Repository) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error WebSocket: %v", err)
		return
	}
	defer conn.Close()

	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			log.Println("Cliente desconectado.")
			break
		}

		var msg MensajeEntrante
		if err := json.Unmarshal(payload, &msg); err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "Formato JSON inválido"}`))
			continue
		}

		// Si el usuario pide crear el mundo, lanzamos una Goroutine para no bloquear
		if msg.Tipo == "INIT_WORLD" {
			go func(m MensajeEntrante) {
				conn.WriteMessage(websocket.TextMessage, []byte(`{"status": "Autenticando perfil en base de datos..."}`))
				
				nivelParaIniciar := 1
				dominioParaIniciar := m.Dominio
				objetivoParaIniciar := m.Objetivo

				// COMPROBACIÓN DE SESIÓN (Login Inteligente)
				userDB, err := database.GetUserProgress(m.Nick)
				if err == nil {
					// Usuario existe, restauramos sus datos
					nivelParaIniciar = userDB.NivelActual
					dominioParaIniciar = userDB.Dominio
					objetivoParaIniciar = userDB.Objetivo
					conn.WriteMessage(websocket.TextMessage, []byte(`{"status": "Sesión encontrada. Restaurando progreso del Consejo..."}`))
				} else {
					// Guardamos al nuevo usuario en Nivel 1
					database.SaveProgress(m.Nick, dominioParaIniciar, objetivoParaIniciar, nivelParaIniciar)
				}

				// Generamos el mundo con el nivel recuperado
				config, err := gemini.GenerateSwarmWorld(context.Background(), m.Nick, m.NivelPrevio, dominioParaIniciar, objetivoParaIniciar, nivelParaIniciar)
				if err != nil {
					log.Printf("Error de LLM: %v", err)
					conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "El Orquestador falló"}`))
					return
				}

				// Convertimos la respuesta nativa a JSON para el Frontend
				respuestaJSON, _ := json.Marshal(map[string]interface{}{
					"tipo": "WORLD_READY",
					"nivel_recuperado": nivelParaIniciar, // Enviamos el nivel recuperado a la UI
					"data": config,
				})
				
				conn.WriteMessage(websocket.TextMessage, respuestaJSON)
			}(msg)
		}

		// Si el usuario envía código para evaluar
		if msg.Tipo == "EVALUATE_CODE" {
			go func(m MensajeEntrante) {
				conn.WriteMessage(websocket.TextMessage, []byte(`{"status": "Auditando código..."}`))
				
				eval, err := gemini.EvaluateAndProgress(context.Background(), m.Nick, m.Dominio, m.Objetivo, m.NivelActual, m.Codigo)
				if err != nil {
					conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "Fallo de evaluación"}`))
					return
				}

				// Hook de Persistencia: Si aprueba, guardamos en base de datos
				if eval.Aprobado {
					errDB := database.SaveProgress(m.Nick, m.Dominio, m.Objetivo, m.NivelActual+1)
					if errDB != nil {
						log.Printf("Aviso: No se pudo guardar progreso: %v", errDB)
					}
				}

				respuestaJSON, _ := json.Marshal(map[string]interface{}{
					"tipo": "CODE_EVALUATED",
					"data": eval,
				})
				
				conn.WriteMessage(websocket.TextMessage, respuestaJSON)
			}(msg)
		}
	}
}
