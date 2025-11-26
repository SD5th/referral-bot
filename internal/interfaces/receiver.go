package interfaces

type UpdateReceiverInterface interface {
	Start() error
	Stop() error
	IsRunning() bool
	GetType() string

	SetUpdateHandler(UpdateHandlerInterface) error
}
