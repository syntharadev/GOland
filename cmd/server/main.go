package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"gemini-go-platform/internal/api"
	"gemini-go-platform/internal/auth"
	"gemini-go-platform/internal/database"
	"gemini-go-platform/internal/llm"
	"gemini-go-platform/internal/mcp"
	"gemini-go-platform/internal/router"

	"github.com/joho/godotenv"
)

func main() {
	// Cargar variables de entorno del archivo .env antes de cualquier inicialización
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: No se pudo cargar el archivo .env, usando variables del sistema.")
	}

	// Inicializar variables de configuración de autenticación y cookie store de forma segura
	auth.Init()

	ctx := context.Background()
	geminiClient, err := llm.InitClient(ctx)
	if err != nil {
		log.Fatalf("Error LLM: %v", err)
	}
	defer geminiClient.Close()

	// Iniciar Servidor MCP MongoDB Atlas
	if err := mcp.IniciarServidorMCPMongo(); err != nil {
		log.Printf("Aviso: No se pudo arrancar el servidor MCP de MongoDB: %v. Operando en modo limitado / simulado.", err)
	} else {
		defer mcp.CerrarServidorMCPMongo()
	}

	// Inicialización de la Base de Datos Supabase (PostgreSQL) - NUNCA SQLite
	dbConn := os.Getenv("DATABASE_URL")
	database, err := database.InitDB(dbConn)
	if err != nil {
		log.Fatalf("Error DB Supabase: %v", err)
	}
	defer database.Close()

	mux := http.NewServeMux()

	// Inyectamos tanto LLM como DB en el handler y lo protegemos con autenticación
	mux.HandleFunc("GET /ws/swarm", auth.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		api.SwarmConnectionHandler(w, r, geminiClient, database)
	}))

	// Rutas de Autenticación OAuth2
	mux.HandleFunc("GET /auth/google/login", auth.HandleGoogleLogin)
	mux.HandleFunc("GET /auth/google/callback", auth.HandleGoogleCallback)
	mux.HandleFunc("GET /auth/status", auth.HandleAuthStatus)
	mux.HandleFunc("GET /auth/logout", auth.HandleLogout)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("GOland operativo."))
	})

	// Servidor de archivos estáticos bajo la ruta web /static/
	staticFS := http.FileServer(http.Dir("./ui/static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", staticFS))

	// Servidor de videos estáticos bajo la ruta web /videos/
	videosFS := http.FileServer(http.Dir("./ui/videos"))
	mux.Handle("GET /videos/", http.StripPrefix("/videos/", videosFS))

	// Ruta raíz (Landing Page con la isla de GOland y el video de intro)
	mux.HandleFunc("GET /", renderIndex)

	// Ruta de inicio de sesión con Google
	mux.HandleFunc("GET /login", renderLogin)

	// Ruta de la aplicación interactiva (Cinemática si no autenticado, Dashboard si autenticado)
	mux.HandleFunc("GET /app", renderApp)

	// Ruta del nuevo Workspace "Pastel Sky" interactivo
	mux.HandleFunc("GET /workspace", renderWorkspace)

	// Endpoint API para chatear con el tutor Gemini (protegido)
	mux.HandleFunc("POST /api/chat", auth.RequireAuth(chatConGeminiHandler))

	// Endpoint API para generar rutas de misiones del enjambre (protegido)
	mux.HandleFunc("POST /api/generar-ruta", auth.RequireAuth(generarRutaHandler))

	// Endpoint Maestro del Orquestador de Enjambre (Swarm Router - protegido)
	mux.HandleFunc("POST /api/orquestador", auth.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		orquestadorHandler(w, r, geminiClient)
	}))

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

// Handler para la landing page (Ruta /)
func renderIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	session, _ := auth.Store.Get(r, "goland-session")
	authVal, ok := session.Values["authenticated"].(bool)
	if ok && authVal {
		http.Redirect(w, r, "/workspace", http.StatusSeeOther)
		return
	}
	http.ServeFile(w, r, "./ui/html/index.html")
}

// Handler para el flujo de la aplicación interactiva (Ruta /app)
func renderApp(w http.ResponseWriter, r *http.Request) {
	session, _ := auth.Store.Get(r, "goland-session")
	authVal, ok := session.Values["authenticated"].(bool)
	if ok && authVal {
		// Redirigir al workspace si ya está autenticado para romper bucles
		http.Redirect(w, r, "/workspace", http.StatusSeeOther)
		return
	}
	// Si no está autenticado, sirve la cinemática de introducción
	http.ServeFile(w, r, "./ui/html/app_GOland.html")
}

// Handler para la pantalla de login (Ruta /login)
func renderLogin(w http.ResponseWriter, r *http.Request) {
	session, _ := auth.Store.Get(r, "goland-session")
	authVal, ok := session.Values["authenticated"].(bool)
	if ok && authVal {
		http.Redirect(w, r, "/workspace", http.StatusSeeOther)
		return
	}
	http.ServeFile(w, r, "./ui/html/login_GOland.html")
}

// Handler para el nuevo Workspace (Ruta /workspace)
func renderWorkspace(w http.ResponseWriter, r *http.Request) {
	session, _ := auth.Store.Get(r, "goland-session")
	authVal, ok := session.Values["authenticated"].(bool)
	if !ok || !authVal {
		// Redirigir a login si no está autenticado
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	http.ServeFile(w, r, "./ui/html/workspace.html")
}

// Estructuras para el chat interactivo de Gemini
type ChatRequest struct {
	Mensaje string `json:"mensaje"`
}

type ChatResponse struct {
	Respuesta string `json:"respuesta"`
}

// Handler para chatear con Gemini (POST /api/chat)
func chatConGeminiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Leer API Key del entorno (Supabase / Gemini Key)
	geminiKey := os.Getenv("GEMINI_API_KEY")
	_ = geminiKey // Ignorar uso por ahora si solo hacemos mock

	var req ChatRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Error decodificando JSON", http.StatusBadRequest)
		return
	}

	// Simulación adaptativa de respuesta del tutor (GOnion / Gemini)
	var respuesta string

	// Generar respuesta simulada adaptativa
	respuesta = "¡Hola! Estoy analizando tu pregunta cuántica: \"" + req.Mensaje + "\". "

	// Mocks inteligentes para demostrar coherencia
	if containsKeyword(req.Mensaje, "variable", "declara", ":=") {
		respuesta += "En Go, declarar una variable con := realiza inferencia automática de tipo y solo está permitida dentro de funciones. Si necesitas declarar una variable a nivel de paquete, debes usar obligatoriamente la sintaxis estándar: var nombre Tipo = valor. ¿Queda clara la diferencia? 🐹⚡"
	} else if containsKeyword(req.Mensaje, "concurrencia", "goroutine", "canal") {
		respuesta += "¡Ah, has tocado el reino cuántico del Cronometrador! Las Goroutines son hilos de ejecución ultraligeros administrados por el runtime de Go. Se inician simplemente anteponiendo la palabra clave 'go' a una llamada a función. Los canales ('chan') son los conductos que permiten a estas goroutines comunicarse de manera segura sin colisiones de memoria. ⏳"
	} else if containsKeyword(req.Mensaje, "sintaxis", "compila", "estatico") {
		respuesta += "El compilador de Go es uno de los más rápidos del mundo. Al ser un lenguaje compilado estáticamente, cualquier error de tipo o asignación incorrecta se detectará al compilar, antes de ejecutar el programa. ¡Esto ahorra horas de debugging en producción! 🛠️"
	} else {
		respuesta += "Esa es una pregunta fascinante. El ecosistema de GOland y tu tutor Gemini están listos para guiarte en el plan de estudios adaptativo de Supabase. Escribe algún fragmento de código Go en el editor de arriba para poner a prueba tu conocimiento. 🚀"
	}

	resp := ChatResponse{
		Respuesta: respuesta,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Función auxiliar para buscar palabras clave de forma case-insensitive básica
func containsKeyword(mensaje string, keywords ...string) bool {
	// Conversión rápida a minúsculas
	mensajeMin := ""
	for _, char := range mensaje {
		if char >= 'A' && char <= 'Z' {
			mensajeMin += string(char + 32)
		} else {
			mensajeMin += string(char)
		}
	}
	for _, kw := range keywords {
		// Búsqueda simple de subcadena
		for i := 0; i <= len(mensajeMin)-len(kw); i++ {
			if mensajeMin[i:i+len(kw)] == kw {
				return true
			}
		}
	}
	return false
}

// Estructuras para la generación de rutas
type RutaRequest struct {
	Proyecto string `json:"proyecto"`
}

type Mision struct {
	Titulo      string `json:"titulo"`
	Concepto    string `json:"concepto"`
	Descripcion string `json:"descripcion"`
}

type RutaResponse struct {
	Misiones []Mision `json:"misiones"`
}

// Handler para generar la ruta de misiones (POST /api/generar-ruta)
func generarRutaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req RutaRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Error decodificando JSON", http.StatusBadRequest)
		return
	}

	// Crear misiones personalizadas dinámicamente basadas en el proyecto
	var misiones []Mision
	proyecto := req.Proyecto
	if proyecto == "" {
		proyecto = "Aplicación General en Go"
	}

	// Mocks inteligentes y hermosos según el proyecto
	misiones = []Mision{
		{
			Titulo:      "Inicialización de Entorno y Variables",
			Concepto:    "Variables e Inferencia Corta (:=)",
			Descripcion: "Establece la estructura básica del proyecto (" + proyecto + ") y configura las variables de entorno necesarias.",
		},
		{
			Titulo:      "Control y Mapeo de Flujo de Datos",
			Concepto:    "Estructuras y Control Flow (if/for/switch)",
			Descripcion: "Crea el motor lógico del proyecto para procesar, iterar y clasificar datos de forma segura.",
		},
		{
			Titulo:      "Ejecución Concurrente del Motor",
			Concepto:    "Concurrencia Cuántica (Goroutines y Canales)",
			Descripcion: "Orquesta tareas en paralelo con goroutines asíncronas para el despacho del proyecto en tiempo real.",
		},
	}

	resp := RutaResponse{
		Misiones: misiones,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Estructuras para el Endpoint Maestro del Orquestador (Swarm Router)
type OrquestadorRequest struct {
	Mensaje string `json:"mensaje"`
	Codigo  string `json:"codigo"`
}

type OrquestadorResponse struct {
	Escuadron int                   `json:"escuadron"`
	Razon     string                `json:"razon"`
	Mensajes  []router.MensajeSwarm `json:"mensajes"`
}

// Handler maestro para enrutar semánticamente y ejecutar los Swarms correspondientes
func orquestadorHandler(w http.ResponseWriter, r *http.Request, geminiClient *llm.GeminiClient) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req OrquestadorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decodificando JSON", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	escuadron, razon, err := router.ClasificarIntencion(ctx, geminiClient, req.Mensaje, req.Codigo)
	if err != nil {
		log.Printf("Aviso: Error en ClasificarIntencion: %v. Usando fallback Escuadrón 1.", err)
		escuadron = 1
		razon = "Interferencia cuántica (fallback automático a Swarm 1)."
	}

	mensajes, err := router.EjecutarSwarm(ctx, geminiClient, escuadron, req.Mensaje, req.Codigo)
	if err != nil {
		log.Printf("Error al ejecutar Swarm %d: %v", escuadron, err)
		http.Error(w, fmt.Sprintf("Error ejecutando el enjambre de agentes: %v", err), http.StatusInternalServerError)
		return
	}

	// Inyectar mensaje inicial de "El Profesor" coordinando el flujo
	introTexto := fmt.Sprintf("¡Excelente consulta! He clasificado tu solicitud y derivado la tarea al **Escuadrón %d** (%s). Dejaré que los especialistas se encarguen de responderte.", escuadron, razon)
	profesorIntro := router.MensajeSwarm{
		Nombre: "El Profesor",
		Texto:  introTexto,
		Avatar: "/static/img/Profesor.png",
	}

	// Prepend
	mensajes = append([]router.MensajeSwarm{profesorIntro}, mensajes...)

	resp := OrquestadorResponse{
		Escuadron: escuadron,
		Razon:     razon,
		Mensajes:  mensajes,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
