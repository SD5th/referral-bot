package bot

import (
	"encoding/json"
	"fmt"
	"os"
	"referral-bot/internal/types"
)

func loadConfig() (*types.BotConfig, error) {
	privateConfigFile, err := os.Open("config/privateConfig.json")
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия privateConfig.json: %v", err)
	}
	defer privateConfigFile.Close()
	var privateConfig *types.BotPrivateConfig
	decoder := json.NewDecoder(privateConfigFile)
	err = decoder.Decode(&privateConfig)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения privateConfig.json: %v", err)
	}

	publicConfigFile, err := os.Open("config/publicConfig.json")
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия publicConfig.json: %v", err)
	}
	defer publicConfigFile.Close()
	var publicConfig *types.BotPublicConfig
	decoder = json.NewDecoder(publicConfigFile)
	err = decoder.Decode(&publicConfig)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения publicConfig.json: %v", err)
	}

	config := types.BotConfig{
		PublicConfig:  publicConfig,
		PrivateConfig: privateConfig,
	}

	if err := verivyConfig(&config); err != nil {
		return nil, fmt.Errorf("Ошибка конфигурации: %v", err)
	}

	return &config, nil
}

func verivyConfig(config *types.BotConfig) error {
	//pubC := config.PublicConfig
	privC := config.PrivateConfig

	if privC.Token == "" {
		return fmt.Errorf("токен не указан в config.json")
	}
	return nil
}
