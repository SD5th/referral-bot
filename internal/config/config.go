package config

type Config struct {
	Bot BotConfig `json:"bot"`
}

type BotConfig struct {
	Debug    bool           `json:"debug"`
	Token    string         `json:"token"`
	Receiver ReceiverConfig `json:"receiver"`
}

type ReceiverConfig struct {
	Type           string   `json:"type"`
	AllowedUpdates []string `json:"allowed_updates"`
}
