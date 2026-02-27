package logx

import "testing"

func TestLogger(t *testing.T) {
	logger, err := New(WithConsole(true))
	if err != nil {
		return
	}
	NewSlog(logger).Info("test")
}
