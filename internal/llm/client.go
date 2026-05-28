package llm

import (
	"context"
	"log"

	"github.com/google/generative-ai-go/genai"
)

type GeminiClient struct {
	Client *genai.Client
}

func InitClient(ctx context.Context) (*GeminiClient, error) {
	log.Println("Mock: Inicializando GeminiClient (Fase 2 no proporcionada explícitamente)")
	return &GeminiClient{
		Client: nil,
	}, nil
}

func (c *GeminiClient) Close() {
	log.Println("Mock: Cerrando GeminiClient")
}

