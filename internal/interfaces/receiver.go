package interfaces

type UpdateReceiver interface {
	Start() error
	Stop() error
	IsRunning() bool
	GetType() string

	SetUpdateHandler(UpdateHandler) error
}
