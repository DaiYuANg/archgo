package logx

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog"
)

// Level 日志级别（强类型）
type Level int

const (
	// TraceLevel 最详细的日志，用于追踪程序执行流程
	TraceLevel Level = iota
	// DebugLevel 调试日志，用于开发环境调试
	DebugLevel
	// InfoLevel 信息日志，记录正常业务流程
	InfoLevel
	// WarnLevel 警告日志，记录潜在问题
	WarnLevel
	// ErrorLevel 错误日志，记录错误但程序仍可运行
	ErrorLevel
	// FatalLevel 致命错误日志，记录后程序会退出
	FatalLevel
	// PanicLevel 恐慌日志，记录后会触发 panic
	PanicLevel
	// DisabledLevel 禁用所有日志
	DisabledLevel
	// NoLevel 无级别（用于特殊场景）
	NoLevel
)

// String 返回日志级别的字符串表示
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

// ParseLevel 从字符串解析日志级别
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

// MustParseLevel 从字符串解析日志级别，失败则 panic
func MustParseLevel(s string) Level {
	level, err := ParseLevel(s)
	if err != nil {
		panic(err)
	}
	return level
}

// ToZerologLevel 转换为 zerolog 级别
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

// Enabled 检查该级别是否启用（比当前级别高的都启用）
func (l Level) Enabled(current Level) bool {
	return l >= current
}

// 预设级别

// Trace 返回 trace 级别
func Trace() Level {
	return TraceLevel
}

// Debug 返回 debug 级别
func Debug() Level {
	return DebugLevel
}

// Info 返回 info 级别
func Info() Level {
	return InfoLevel
}

// Warn 返回 warn 级别
func Warn() Level {
	return WarnLevel
}

// Error 返回 error 级别
func Error() Level {
	return ErrorLevel
}

// Fatal 返回 fatal 级别
func Fatal() Level {
	return FatalLevel
}

// Panic 返回 panic 级别
func Panic() Level {
	return PanicLevel
}

// Disabled 返回 disabled 级别
func Disabled() Level {
	return DisabledLevel
}
