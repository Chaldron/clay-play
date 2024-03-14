package logger

import (
	"fmt"
	"log"
	"os"
)

type Logger interface {
	Printf(string, ...any)
	Errorf(string, ...any)
	Fatalf(string, ...any)
}

type StdLogger struct {
	log *log.Logger
}

func NewStdLogger() *StdLogger {
	l := log.New(os.Stdout, "", log.LstdFlags)
	return &StdLogger{
		log: l,
	}
}

func (l *StdLogger) Printf(format string, v ...any) {
	l.log.Printf(format, v...)
}

func (l *StdLogger) Errorf(format string, v ...any) {
	l.log.Printf("ERROR %s", fmt.Sprintf(format, v...))
}

func (l *StdLogger) Fatalf(format string, v ...any) {
	l.log.Fatalf(format, v...)
}

type NoopLogger struct{}

func NewNoopLogger() *NoopLogger {
	return &NoopLogger{}
}

func (l *NoopLogger) Printf(format string, v ...any) {}
func (l *NoopLogger) Errorf(format string, v ...any) {}
func (l *NoopLogger) Fatalf(format string, v ...any) {
	os.Exit(1)
}
