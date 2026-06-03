package config

import (
	"os"
)

// GetGeminiAPIKey recupera la clave de API de Gemini de las variables de entorno
func GetGeminiAPIKey() string {
	return os.Getenv("GEMINI_API_KEY")
}

// GetOptiLLMEndpoint recupera la URL base para el proxy de OptiLLM, si está configurado
func GetOptiLLMEndpoint() string {
	endpoint := os.Getenv("OPTILLM_ENDPOINT")
	if endpoint == "" {
		// Retornar vacío para que use el endpoint oficial si no se especifica
		return ""
	}
	return endpoint
}

// GetMongoDBURI recupera la URI de conexión a MongoDB Atlas
func GetMongoDBURI() string {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		return "mongodb://localhost:27017"
	}
	return uri
}
