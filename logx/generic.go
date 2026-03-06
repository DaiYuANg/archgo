package logx

import "github.com/samber/lo"

// WithFieldT adds one typed field to logger and returns derived logger.
func WithFieldT[T any](logger *Logger, key string, value T) *Logger {
	if logger == nil {
		return nil
	}
	return logger.WithField(key, value)
}

// WithFieldsT adds typed fields to logger and returns derived logger.
func WithFieldsT[T any](logger *Logger, fields map[string]T) *Logger {
	if logger == nil {
		return nil
	}
	converted := lo.MapValues(fields, func(value T, _ string) interface{} {
		return value
	})
	return logger.WithFields(converted)
}
