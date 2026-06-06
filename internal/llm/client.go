package llm

import (
	"context"
	"log"
	"sync"

	"gemini-go-platform/internal/config"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiClient struct {
	Client       *genai.Client
	ChatSessions sync.Map
}

func InitClient(ctx context.Context) (*GeminiClient, error) {
	apiKey := config.GetGeminiAPIKey()
	if apiKey == "" {
		log.Println("Aviso: GEMINI_API_KEY no configurada. Operando en modo de simulación (MOCK).")
		return &GeminiClient{
			Client: nil,
		}, nil
	}

	var opts []option.ClientOption
	opts = append(opts, option.WithAPIKey(apiKey))

	endpoint := config.GetOptiLLMEndpoint()
	if endpoint != "" {
		log.Printf("Conectando GeminiClient al proxy de OptiLLM: %s", endpoint)
		opts = append(opts, option.WithEndpoint(endpoint))
	} else {
		log.Println("Conectando GeminiClient directamente al servidor oficial de Google.")
	}

	client, err := genai.NewClient(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return &GeminiClient{
		Client: client,
	}, nil
}

func (c *GeminiClient) Close() {
	if c.Client != nil {
		c.Client.Close()
		log.Println("Cliente de Gemini cerrado correctamente.")
	}
}
