package bot

type UpdateReceiver interface {
	Start() error

	Stop() error

	IsRunning() bool

	GetType() string
}
