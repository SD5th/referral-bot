package bot

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Token              string `json:"token"`
	Debug              bool   `json:"debug"`
	UpdateReceiverType string `json:"updateReceiverType"`
}

func loadConfig() (*Config, error) {
	file, err := os.Open("config.json")
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия config.json: %v", err)
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения config.json: %v", err)
	}

	if config.Token == "" {
		return nil, fmt.Errorf("токен не указан в config.json")
	}

	return &config, nil
}
