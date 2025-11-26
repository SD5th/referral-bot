package interfaces

type LoggerInterface interface {
	Info(format string, v ...any)
	Error(format string, v ...any)
	Warn(format string, v ...any)
	Fatal(format string, v ...any)
}
