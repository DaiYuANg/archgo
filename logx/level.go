package logx

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog"
)

// Level documents related behavior.
type Level int

const (
	// TraceLevel documents related behavior.
	TraceLevel Level = iota
	// DebugLevel documents related behavior.
	DebugLevel
	// InfoLevel logs related events.
	InfoLevel
	// WarnLevel logs related events.
	WarnLevel
	// ErrorLevel logs related events.
	ErrorLevel
	// FatalLevel logs related events.
	FatalLevel
	// PanicLevel logs related events.
	PanicLevel
	// DisabledLevel disables related functionality.
	DisabledLevel
	// NoLevel documents related behavior.
	NoLevel
)

// String returns related data.
func (l Level) String() string {
	switch l {
	case TraceLevel:
		return "trace"
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	case PanicLevel:
		return "panic"
	case DisabledLevel:
		return "disabled"
	case NoLevel:
		return "none"
	default:
		return "unknown"
	}
}

// ParseLevel parses related input.
func ParseLevel(s string) (Level, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "trace":
		return TraceLevel, nil
	case "debug":
		return DebugLevel, nil
	case "info":
		return InfoLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
	case "error":
		return ErrorLevel, nil
	case "fatal":
		return FatalLevel, nil
	case "panic":
		return PanicLevel, nil
	case "disabled", "disable":
		return DisabledLevel, nil
	case "none", "no", "":
		return NoLevel, nil
	default:
		return NoLevel, fmt.Errorf("invalid log level: %s", s)
	}
}

// MustParseLevel parses related input.
func MustParseLevel(s string) Level {
	level, err := ParseLevel(s)
	if err != nil {
		panic(err)
	}
	return level
}

// ToZerologLevel converts related values.
func (l Level) ToZerologLevel() zerolog.Level {
	switch l {
	case TraceLevel:
		return zerolog.TraceLevel
	case DebugLevel:
		return zerolog.DebugLevel
	case InfoLevel:
		return zerolog.InfoLevel
	case WarnLevel:
		return zerolog.WarnLevel
	case ErrorLevel:
		return zerolog.ErrorLevel
	case FatalLevel:
		return zerolog.FatalLevel
	case PanicLevel:
		return zerolog.PanicLevel
	case DisabledLevel:
		return zerolog.Disabled
	case NoLevel:
		return zerolog.NoLevel
	default:
		return zerolog.InfoLevel
	}
}

// Enabled checks related state.
func (l Level) Enabled(current Level) bool {
	return l >= current
}

// Note.

// Trace returns related data.
func Trace() Level {
	return TraceLevel
}

// Debug returns related data.
func Debug() Level {
	return DebugLevel
}

// Info returns related data.
func Info() Level {
	return InfoLevel
}

// Warn returns related data.
func Warn() Level {
	return WarnLevel
}

// Error returns related data.
func Error() Level {
	return ErrorLevel
}

// Fatal returns related data.
func Fatal() Level {
	return FatalLevel
}

// Panic returns related data.
func Panic() Level {
	return PanicLevel
}

// Disabled returns related data.
func Disabled() Level {
	return DisabledLevel
}
