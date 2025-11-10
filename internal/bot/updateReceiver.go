package bot

type UpdateReceiver interface {
	start() error
	stop() error
	isRunning() bool
	getType() string
}
