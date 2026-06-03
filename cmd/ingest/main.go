package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/api/option"
)

type GoDoc struct {
	Title     string    `bson:"title"`
	Content   string    `bson:"content"`
	Embedding []float32 `bson:"embedding"`
	Source    string    `bson:"source"`
	UpdatedAt time.Time `bson:"updated_at"`
}

func main() {
	log.Println("=== Iniciando Motor de Ingesta Vectorial para GOland ===")

	// 1. Cargar clave de API y URI de MongoDB
	geminiKey := os.Getenv("GEMINI_API_KEY")
	if geminiKey == "" {
		log.Fatal("Error: La variable de entorno GEMINI_API_KEY es requerida para generar embeddings.")
	}

	mongodbURI := os.Getenv("MONGODB_URI")
	if mongodbURI == "" {
		mongodbURI = "mongodb://localhost:27017"
		log.Printf("Aviso: MONGODB_URI no está configurada, usando default: %s", mongodbURI)
	}

	docsDir := "./docs/go_oficial"
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		log.Fatalf("Error creando el directorio de documentación: %v", err)
	}

	// Crear algunos archivos Markdown de documentación si no hay ninguno
	crearDocumentosDeEjemplo(docsDir)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// 2. Inicializar Cliente MongoDB
	log.Println("Conectando a MongoDB Atlas...")
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongodbURI))
	if err != nil {
		log.Fatalf("Error al conectar a MongoDB: %v", err)
	}
	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Printf("Error al cerrar conexión con MongoDB: %v", err)
		}
	}()

	db := mongoClient.Database("goland_db")
	coll := db.Collection("go_docs")

	// 3. Inicializar Cliente Gemini para embeddings
	log.Println("Inicializando cliente de Google AI (EmbeddingModel)...")
	genClient, err := genai.NewClient(ctx, option.WithAPIKey(geminiKey))
	if err != nil {
		log.Fatalf("Error inicializando cliente de Gemini: %v", err)
	}
	defer genClient.Close()
	embedModel := genClient.EmbeddingModel("text-embedding-004")

	// 4. Leer archivos Markdown e indexar
	log.Printf("Buscando archivos Markdown en '%s'...", docsDir)
	var filesProcessed int
	err = filepath.WalkDir(docsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}

		log.Printf("Procesando: %s", d.Name())
		contentBytes, err := os.ReadFile(path)
		if err != nil {
			log.Printf("Error leyendo archivo %s: %v", path, err)
			return nil
		}
		content := string(contentBytes)
		title := strings.TrimSuffix(d.Name(), ".md")

		// Generar Embedding vectorial
		res, err := embedModel.EmbedContent(ctx, genai.Text(content))
		if err != nil {
			log.Printf("Error generando embedding para %s: %v", d.Name(), err)
			return nil
		}

		if res.Embedding == nil || len(res.Embedding.Values) == 0 {
			log.Printf("Warning: Embedding retornado vacío para %s", d.Name())
			return nil
		}

		// Crear documento para guardar en MongoDB
		doc := GoDoc{
			Title:     title,
			Content:   content,
			Embedding: res.Embedding.Values,
			Source:    fmt.Sprintf("docs/go_oficial/%s", d.Name()),
			UpdatedAt: time.Now(),
		}

		// Insertar o actualizar usando upsert
		filter := bson.M{"source": doc.Source}
		update := bson.M{"$set": doc}
		opts := options.Update().SetUpsert(true)

		_, err = coll.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			log.Printf("Error insertando documento %s en MongoDB: %v", d.Name(), err)
			return nil
		}

		log.Printf("✓ Documento '%s' indexado exitosamente en MongoDB (Vector de %d dimensiones).", title, len(doc.Embedding))
		filesProcessed++
		return nil
	})

	if err != nil {
		log.Fatalf("Error recorriendo el directorio de documentación: %v", err)
	}

	log.Printf("=== Ingesta Vectorial Completada. %d archivos procesados. ===", filesProcessed)
}

func crearDocumentosDeEjemplo(dir string) {
	goroutinesPath := filepath.Join(dir, "Goroutines.md")
	if _, err := os.Stat(goroutinesPath); os.IsNotExist(err) {
		content := `# Fundamentos de Concurrencia: Goroutines
Una goroutine es un hilo de ejecución ligero administrado por el runtime de Go.
Sintaxis: go f(x, y, z) arranca una nueva goroutine que ejecuta f.
Las goroutines comparten el mismo espacio de direcciones, por lo que el acceso a la memoria compartida debe ser sincronizado utilizando canales o el paquete sync de la biblioteca estándar.`
		_ = os.WriteFile(goroutinesPath, []byte(content), 0644)
	}

	channelsPath := filepath.Join(dir, "Canales.md")
	if _, err := os.Stat(channelsPath); os.IsNotExist(err) {
		content := `# Canales de Comunicación (Channels)
Los canales son los conductos tipados a través de los cuales puedes enviar y recibir valores con el operador de canal, <-.
Se inicializan usando make: ch := make(chan int).
Por defecto, los envíos y recepciones se bloquean hasta que el otro lado esté listo, lo que permite sincronizar goroutines de forma nativa sin bloqueos explícitos ni variables de condición.`
		_ = os.WriteFile(channelsPath, []byte(content), 0644)
	}
}
