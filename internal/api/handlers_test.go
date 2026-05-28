package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"gemini-go-platform/internal/db"
	"gemini-go-platform/internal/llm"
	"github.com/gorilla/websocket"
)

func TestSwarmConnectionHandler(t *testing.T) {
	// Create mock gemini client
	geminiClient, _ := llm.InitClient(context.Background())
	database, _ := db.InitDB("file::memory:?cache=shared")
	defer database.Close()
	
	// Create a test server with our handler wrapper
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		SwarmConnectionHandler(w, r, geminiClient, database)
	}))
	defer s.Close()

	// Convert http:// to ws:// URL
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// Connect to the server
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("Failed to connect to websocket: %v", err)
	}
	defer ws.Close()

	// Send an INIT_WORLD message to the Swarm connection
	msgObj := MensajeEntrante{
		Tipo:     "INIT_WORLD",
		Dominio:  "Test",
		Objetivo: "Test Objetivo",
	}
	msg, _ := json.Marshal(msgObj)
	
	if err := ws.WriteMessage(websocket.TextMessage, msg); err != nil {
		t.Fatalf("Failed to write message: %v", err)
	}

	// Read the expected response (Orquestador analizando...)
	_, p1, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read message 1: %v", err)
	}

	expected1 := `{"status": "Autenticando perfil en base de datos..."}`
	if string(p1) != expected1 {
		t.Errorf("Expected message %q, got %q", expected1, string(p1))
	}
	
	// Read the second response (WORLD_READY)
	_, p2, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read message 2: %v", err)
	}
	
	if !strings.Contains(string(p2), "WORLD_READY") {
		t.Errorf("Expected WORLD_READY message, got %q", string(p2))
	}
}
