package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"gemini-go-platform/internal/api"
	"gemini-go-platform/internal/auth"
	"gemini-go-platform/internal/database"
	"gemini-go-platform/internal/llm"
)

func main() {
	ctx := context.Background()
	geminiClient, err := llm.InitClient(ctx)
	if err != nil {
		log.Fatalf("Error LLM: %v", err)
	}
	defer geminiClient.Close()

	// Inicialización de la Base de Datos Supabase (PostgreSQL) - NUNCA SQLite
	dbConn := os.Getenv("DATABASE_URL")
	database, err := database.InitDB(dbConn)
	if err != nil {
		log.Fatalf("Error DB Supabase: %v", err)
	}
	defer database.Close()

	mux := http.NewServeMux()

	// Inyectamos tanto LLM como DB en el handler
	mux.HandleFunc("GET /ws/swarm", func(w http.ResponseWriter, r *http.Request) {
		api.SwarmConnectionHandler(w, r, geminiClient, database)
	})
	
	// Rutas de Autenticación OAuth2
	mux.HandleFunc("GET /auth/google/login", auth.HandleGoogleLogin)
	mux.HandleFunc("GET /auth/google/callback", auth.HandleGoogleCallback)
	mux.HandleFunc("GET /auth/status", auth.HandleAuthStatus)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("GOland operativo."))
	})

	// Servidor de archivos estáticos bajo la ruta web /static/
	staticFS := http.FileServer(http.Dir("./ui/static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", staticFS))

	// Ruta raíz que sirve el archivo index.html desde su nueva ubicación
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "./ui/html/index.html")
	})

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Println("🚀 Servidor en http://localhost:8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Fallo en servidor: %v", err)
	}
}
