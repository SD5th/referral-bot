package logger

import (
	"log"
	"os"
)

type StdLogger struct {
	logger *log.Logger
}

func NewStdLogger() *StdLogger {
	return &StdLogger{
		logger: log.New(os.Stdout, "[Bot] ", log.Ldate|log.Ltime),
	}
}

func (l *StdLogger) Info(format string, v ...any) {
	l.logger.Printf("[INFO] "+format, v...)
}

func (l *StdLogger) Warn(format string, v ...any) {
	l.logger.Printf("[WARN] "+format, v...)
}

func (l *StdLogger) Error(format string, v ...any) {
	l.logger.Printf("[ERROR] "+format, v...)
}

func (l *StdLogger) Fatal(format string, v ...any) {
	l.logger.Printf("[FATAL] "+format, v...)
	os.Exit(1)
}
