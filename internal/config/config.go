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
	Type           string        `json:"type"`
	AllowedUpdates []string      `json:"allowed_updates"`
	BufferSize     int           `json:"buffer_size"`
	Webhook        WebhookConfig `json:"webhook"`
	Poller         PollerConfig  `json:"poller"`
}

type WebhookConfig struct {
	ServerIP    string `json:"server_ip"`
	ListenIP    string `json:"listen_ip" default:"0.0.0.0"`
	Port        string `json:"port" default:"8443"`
	CertFile    string `json:"cert_file"`
	KeyFile     string `json:"key_file"`
	ReadTimeout int    `json:"read_timeout"`
}

type PollerConfig struct {
	Offset  int `json:"offset"`
	Timeout int `json:"timeout"`
}
