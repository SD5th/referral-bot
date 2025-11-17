package config

import (
	"encoding/json"
	"fmt"
	"os"
)

func LoadConfig() (*Config, error) {
	var config *Config

	privateConfigFile, err := os.Open("config/privateConfig.json")
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия privateConfig.json: %v", err)
	}
	defer privateConfigFile.Close()

	decoder := json.NewDecoder(privateConfigFile)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения privateConfig.json: %v", err)
	}

	publicConfigFile, err := os.Open("config/publicConfig.json")
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия publicConfig.json: %v", err)
	}
	defer publicConfigFile.Close()
	decoder = json.NewDecoder(publicConfigFile)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения publicConfig.json: %v", err)
	}

	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("Ошибка конфигурации: %v", err)
	}

	return config, nil
}
