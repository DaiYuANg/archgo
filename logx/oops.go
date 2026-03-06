package logx

import (
	"context"
	"fmt"

	"github.com/samber/oops"
)

// LogOops logs related events.
func (l *Logger) LogOops(err error) {
	if l == nil {
		return
	}

	// Note.
	l.Error("Error", "error", err)
}

// LogOopsWithStack logs related events.
func (l *Logger) LogOopsWithStack(err error) {
	l.LogOops(err)
}

// Oops creates related functionality.
func (l *Logger) Oops() error {
	return oops.New("error")
}

// Oopsf creates related functionality.
func (l *Logger) Oopsf(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	if msg == "" {
		msg = "error"
	}
	return oops.New(msg)
}

// OopsWith creates related functionality.
func (l *Logger) OopsWith(ctx context.Context) error {
	_ = ctx
	return oops.New("error")
}
