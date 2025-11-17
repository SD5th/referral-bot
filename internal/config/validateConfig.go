package config

import "fmt"

func validateConfig(config *Config) error {
	if config.Bot.Token == "" {
		return fmt.Errorf("токен не указан в config.json")
	}
	return nil
}
