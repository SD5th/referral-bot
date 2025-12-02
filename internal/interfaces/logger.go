package interfaces

type Logger interface {
	Warn(format string, v ...any)
	Info(format string, v ...any)
	Error(format string, v ...any)
	Fatal(format string, v ...any)
}
