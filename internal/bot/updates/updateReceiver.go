package updates

type UpdateReceiver interface {
	Start() error
	Stop() error
	IsRunning() bool
	GetType() string
}

type UpdateReceiverConfig struct {
	AllowedUpdates []string
}
