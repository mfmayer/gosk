package utils

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()
}

// GetOpenAIKey tries to retrieve OpenAI key from "OPENAI_API_KEY" environment variable or .env file in current workgin directory
func GetOpenAIKey() (string, error) {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		return "", errors.New("openai api key not found")
	}
	return key, nil
}
