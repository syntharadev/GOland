package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int64          `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      *int64      `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type InitializeResult struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ServerInfo      ServerInfo             `json:"serverInfo"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ToolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type AggregateArgs struct {
	DB         string        `json:"db"`
	Collection string        `json:"collection"`
	Pipeline   []interface{} `json:"pipeline"`
}

type ToolCallResult struct {
	Content []TextContent `json:"content"`
}

type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func main() {
	// Redirigir el logger de Go standard a Stderr para no contaminar Stdout (JSON-RPC)
	log.SetOutput(os.Stderr)
	log.Println("Arrancando Servidor MCP Nativo de MongoDB en Go...")

	mongodbURI := os.Getenv("MONGODB_URI")
	if mongodbURI == "" {
		mongodbURI = "mongodb://localhost:27017"
		log.Printf("MONGODB_URI vacía. Usando valor por defecto: %s", mongodbURI)
	}

	// Conectarse a MongoDB Atlas
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongodbURI))
	if err != nil {
		log.Fatalf("Error crítico al conectar a MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())

	// Verificar conexión
	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Fallo en el ping a MongoDB: %v", err)
	}
	log.Println("✓ Conectado correctamente a MongoDB Atlas")

	// Leer stdin línea por línea
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("Stdin cerrado (EOF). Deteniendo servidor MCP.")
				break
			}
			log.Printf("Error leyendo línea de stdin: %v", err)
			continue
		}

		// Procesar petición JSON-RPC
		go procesarPeticion(line, mongoClient)
	}
}

func procesarPeticion(payload []byte, client *mongo.Client) {
	var req JSONRPCRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		enviarError(nil, -32700, "Parse error: "+err.Error())
		return
	}

	switch req.Method {
	case "initialize":
		res := InitializeResult{
			ProtocolVersion: "2024-11-05",
			Capabilities: map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			ServerInfo: ServerInfo{
				Name:    "mcp-mongo-go",
				Version: "1.0.0",
			},
		}
		enviarRespuesta(req.ID, res)

	case "notifications/initialized":
		// No requiere respuesta
		log.Println("MCP Handshake completado exitosamente.")

	case "tools/list":
		// Responder con la herramienta que ofrecemos
		tools := []map[string]interface{}{
			{
				"name":        "aggregate",
				"description": "Ejecuta un pipeline de agregación en la base de datos de MongoDB. Se usa para realizar búsquedas vectoriales con $vectorSearch.",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"db":         map[string]string{"type": "string"},
						"collection": map[string]string{"type": "string"},
						"pipeline":   map[string]string{"type": "string"},
					},
					"required": []string{"db", "collection", "pipeline"},
				},
			},
		}
		enviarRespuesta(req.ID, map[string]interface{}{
			"tools": tools,
		})

	case "tools/call":
		var params ToolCallParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			enviarError(req.ID, -32602, "Invalid params: "+err.Error())
			return
		}

		if params.Name != "aggregate" {
			enviarError(req.ID, -32601, fmt.Sprintf("Method not found: tool '%s' no soportada", params.Name))
			return
		}

		var args AggregateArgs
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			enviarError(req.ID, -32602, "Error decodificando argumentos de agregación: "+err.Error())
			return
		}

		if args.DB == "" {
			args.DB = "goland_db"
		}
		if args.Collection == "" {
			args.Collection = "go_docs"
		}

		log.Printf("Ejecutando pipeline de agregación en %s.%s", args.DB, args.Collection)

		// Ejecutar la agregación en MongoDB
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		coll := client.Database(args.DB).Collection(args.Collection)
		cursor, err := coll.Aggregate(ctx, args.Pipeline)
		if err != nil {
			enviarError(req.ID, -32603, "Error ejecutando agregación: "+err.Error())
			return
		}
		defer cursor.Close(ctx)

		var resultados []bson.M
		if err := cursor.All(ctx, &resultados); err != nil {
			enviarError(req.ID, -32603, "Error decodificando resultados: "+err.Error())
			return
		}

		jsonData, err := json.Marshal(resultados)
		if err != nil {
			enviarError(req.ID, -32603, "Error serializando resultados a JSON: "+err.Error())
			return
		}

		// Responder en formato compatible con el estándar MCP (content)
		mcpRes := ToolCallResult{
			Content: []TextContent{
				{
					Type: "text",
					Text: string(jsonData),
				},
			},
		}

		enviarRespuesta(req.ID, mcpRes)

	default:
		if req.ID != nil {
			enviarError(req.ID, -32601, fmt.Sprintf("Method not found: %s", req.Method))
		}
	}
}

func enviarRespuesta(id *int64, result interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error serializando respuesta JSON-RPC: %v", err)
		return
	}

	os.Stdout.Write(append(data, '\n'))
}

func enviarError(id *int64, code int, message string) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error serializando error JSON-RPC: %v", err)
		return
	}

	os.Stdout.Write(append(data, '\n'))
}
