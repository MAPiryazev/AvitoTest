package config

import (
	"fmt"
	"os"

	customerrors "AvitoTest/internal/custom-errors"
)

type APIConfig struct {
	Port string
}

// LoadAPIConfig — загружает конфиг API
func LoadAPIConfig() (*APIConfig, error) {
	port := os.Getenv("API_PORT")
	if port == "" {
		return nil, fmt.Errorf("ошибка конфигурации API: %w", customerrors.ErrParamNotFound)
	}

	return &APIConfig{
		Port: port,
	}, nil
}
