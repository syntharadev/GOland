package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
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
	// Cargar variables de entorno del archivo .env antes de cualquier inicialización
	_ = godotenv.Load()

	log.Println("=== Iniciando Motor de Ingesta Vectorial Oficial para GOland ===")

	// 1. Cargar configuración
	geminiKey := os.Getenv("GEMINI_API_KEY")
	if geminiKey == "" {
		log.Println("Aviso: GEMINI_API_KEY no configurada. Se utilizarán vectores mock simulados.")
	}

	mongodbURI := os.Getenv("MONGODB_URI")
	if mongodbURI == "" {
		mongodbURI = "mongodb://localhost:27017"
		log.Printf("Aviso: MONGODB_URI no está configurada, usando default: %s", mongodbURI)
	}

	// 2. Clonación Temporal del Repositorio de Go
	tempDir := "./tmp_go_docs"
	log.Println("Fase 1: Clonación del repositorio oficial de Go...")
	if _, err := os.Stat(tempDir); err == nil {
		log.Printf("Borrando carpeta temporal anterior '%s'...", tempDir)
		if err := os.RemoveAll(tempDir); err != nil {
			log.Fatalf("Error al limpiar carpeta temporal: %v", err)
		}
	}

	// Ejecutar clonación superficial (shallow clone)
	cmdClone := exec.Command("git", "clone", "--depth", "1", "https://github.com/golang/go.git", tempDir)
	cmdClone.Stdout = os.Stdout
	cmdClone.Stderr = os.Stderr
	log.Println("Ejecutando: git clone --depth 1 https://github.com/golang/go.git ./tmp_go_docs")
	if err := cmdClone.Run(); err != nil {
		log.Fatalf("Error al clonar el repositorio de Go: %v", err)
	}

	// Asegurar limpieza de la carpeta al finalizar la ejecución
	defer func() {
		log.Printf("Limpiando carpeta temporal '%s'...", tempDir)
		if err := os.RemoveAll(tempDir); err != nil {
			log.Printf("Error al limpiar carpeta temporal al salir: %v", err)
		} else {
			log.Println("✓ Limpieza completada con éxito.")
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	// 3. Inicializar Cliente MongoDB
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

	// 4. Inicializar Cliente Gemini
	var genClient *genai.Client
	var embedModel *genai.EmbeddingModel
	if geminiKey != "" {
		log.Println("Conectando directamente a Google API para Embeddings (Bypass de OptiLLM)...")
		genClient, err = genai.NewClient(ctx, option.WithAPIKey(geminiKey))
		if err != nil {
			log.Fatalf("Error inicializando cliente de Gemini: %v", err)
		}
		defer genClient.Close()
		embedModel = genClient.EmbeddingModel("text-embedding-004")
	}

	// 5. Explorar recursivamente las carpetas doc y src
	dirsToExplore := []string{
		filepath.Join(tempDir, "doc"),
		filepath.Join(tempDir, "src"),
	}

	log.Println("Fase 2: Extracción y Limpieza de datos en marcha...")
	var totalProcessed int

	for _, docsDir := range dirsToExplore {
		if _, err := os.Stat(docsDir); os.IsNotExist(err) {
			log.Printf("Aviso: El directorio '%s' no existe, omitiendo.", docsDir)
			continue
		}

		log.Printf("Analizando directorio: %s", docsDir)
		err = filepath.WalkDir(docsDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			// Omitir directorios
			if d.IsDir() {
				return nil
			}

			// Ignorar carpetas de testdata y vendor
			pathLower := strings.ToLower(path)
			if strings.Contains(pathLower, "testdata") || strings.Contains(pathLower, "vendor") {
				return nil
			}

			// Filtrar por extensiones válidas (.md y .html)
			ext := strings.ToLower(filepath.Ext(d.Name()))
			if ext != ".md" && ext != ".html" {
				return nil
			}

			log.Printf("Procesando archivo: %s...", path)
			contentBytes, err := os.ReadFile(path)
			if err != nil {
				log.Printf("Error leyendo archivo %s: %v", path, err)
				return nil
			}

			content := string(contentBytes)
			title := strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))

			// Fase 3: Fragmentación (Chunking) en bloques de ~1000 caracteres
			chunks := chunkText(content, 1000)
			log.Printf("-> Dividido en %d fragmentos (chunks). Indexando en la base de datos...", len(chunks))

			for idx, chunk := range chunks {
				// Evitar sobrecargar la API
				time.Sleep(80 * time.Millisecond)

				var embeddingValues []float32
				if embedModel != nil {
					res, err := embedModel.EmbedContent(ctx, genai.Text(chunk))
					if err != nil {
						log.Printf("Error generando embedding para %s (chunk %d): %v. Usando vector mock...", title, idx, err)
						embeddingValues = getMockEmbedding()
					} else if res.Embedding == nil || len(res.Embedding.Values) == 0 {
						log.Printf("Warning: Embedding vacío para %s (chunk %d). Usando vector mock...", title, idx)
						embeddingValues = getMockEmbedding()
					} else {
						embeddingValues = res.Embedding.Values
					}
				} else {
					embeddingValues = getMockEmbedding()
				}

				// Crear documento del fragmento
				docSource := fmt.Sprintf("%s#chunk-%d", filepath.ToSlash(path), idx)
				docTitle := fmt.Sprintf("%s (Parte %d)", title, idx+1)

				doc := GoDoc{
					Title:     docTitle,
					Content:   chunk,
					Embedding: embeddingValues,
					Source:    docSource,
					UpdatedAt: time.Now(),
				}

				// Guardar en MongoDB Atlas mediante Upsert
				filter := bson.M{"source": doc.Source}
				update := bson.M{"$set": doc}
				opts := options.Update().SetUpsert(true)

				_, err = coll.UpdateOne(ctx, filter, update, opts)
				if err != nil {
					log.Printf("Error insertando chunk %d de %s en MongoDB: %v", idx, title, err)
					continue
				}
				totalProcessed++
			}

			log.Printf("✓ Archivo completo '%s' procesado exitosamente.", title)
			return nil
		})

		if err != nil {
			log.Printf("Error recorriendo el directorio %s: %v", docsDir, err)
		}
	}

	log.Printf("=== Ingesta Vectorial Completada. %d chunks indexados. ===", totalProcessed)
}

// chunkText divide el texto en bloques de chunkSize intentando no romper oraciones.
func chunkText(text string, chunkSize int) []string {
	var chunks []string
	runes := []rune(text)
	length := len(runes)

	if length <= chunkSize {
		return []string{text}
	}

	start := 0
	for start < length {
		end := start + chunkSize
		if end >= length {
			chunks = append(chunks, string(runes[start:]))
			break
		}

		// Intentar retroceder hasta encontrar un signo de puntuación o salto de línea
		bestEnd := end
		lookbackLimit := end - (chunkSize / 8) // no retroceder más del 12.5% del tamaño del chunk
		if lookbackLimit < start {
			lookbackLimit = start
		}

		found := false
		for i := end; i > lookbackLimit; i-- {
			r := runes[i]
			if r == '.' || r == '?' || r == '!' || r == '\n' {
				bestEnd = i + 1
				found = true
				break
			}
		}

		// Si no se encuentra puntuación, buscar un espacio
		if !found {
			for i := end; i > lookbackLimit; i-- {
				if runes[i] == ' ' {
					bestEnd = i + 1
					found = true
					break
				}
			}
		}

		chunks = append(chunks, string(runes[start:bestEnd]))
		start = bestEnd

		// Prevenir bucles infinitos si no hay avance
		if start == end && !found {
			start = end
		}
	}

	return chunks
}

// getMockEmbedding genera un vector mock de 768 dimensiones.
func getMockEmbedding() []float32 {
	mockVector := make([]float32, 768)
	for i := range mockVector {
		mockVector[i] = 0.05
	}
	return mockVector
}
