package logx

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Option documents related behavior.
type Option func(*config)

// config documents related behavior.
type config struct {
	level      Level
	console    bool
	noColor    bool
	filePath   string
	maxSize    int
	maxAge     int
	maxBackups int
	timeFormat string
	setGlobal  bool
	addCaller  bool
	localTime  bool
	compress   bool
}

// defaultConfig provides default behavior.
func defaultConfig() config {
	return config{
		level:      InfoLevel,
		console:    true,
		noColor:    false,
		timeFormat: "2006-01-02 15:04:05",
		maxSize:    100,  // 100MB
		maxAge:     7,    // 7 days
		maxBackups: 10,   // 10 files
		localTime:  true, // use local time for rotation
		compress:   true, // compress rotated files
	}
}

// validate documents related behavior.
func (c *config) validate() error {
	// Note.
	if c.level < TraceLevel || c.level > DisabledLevel {
		return fmt.Errorf("invalid log level: %v", c.level)
	}

	// Note.
	if c.filePath != "" {
		// Note.
		dir := filepath.Dir(c.filePath)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			// Note.
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("cannot create log directory %s: %w", dir, err)
			}
		}
	}

	// Note.
	if c.maxSize < 1 {
		return fmt.Errorf("maxSize must be at least 1MB, got %d", c.maxSize)
	}

	// Note.
	if c.maxAge < 0 {
		return fmt.Errorf("maxAge cannot be negative, got %d", c.maxAge)
	}

	// Note.
	if c.maxBackups < 0 {
		return fmt.Errorf("maxBackups cannot be negative, got %d", c.maxBackups)
	}

	return nil
}

// Note.

// WithLevel configures related behavior.
// Note.
func WithLevel(level Level) Option {
	return func(c *config) {
		c.level = level
	}
}

// WithLevelString configures related behavior.
func WithLevelString(level string) Option {
	return func(c *config) {
		l, err := ParseLevel(level)
		if err == nil {
			c.level = l
		}
	}
}

// WithTraceLevel enables related functionality.
func WithTraceLevel() Option {
	return WithLevel(TraceLevel)
}

// WithDebugLevel enables related functionality.
func WithDebugLevel() Option {
	return WithLevel(DebugLevel)
}

// WithInfoLevel enables related functionality.
func WithInfoLevel() Option {
	return WithLevel(InfoLevel)
}

// WithWarnLevel enables related functionality.
func WithWarnLevel() Option {
	return WithLevel(WarnLevel)
}

// WithErrorLevel enables related functionality.
func WithErrorLevel() Option {
	return WithLevel(ErrorLevel)
}

// WithFatalLevel enables related functionality.
func WithFatalLevel() Option {
	return WithLevel(FatalLevel)
}

// WithPanicLevel enables related functionality.
func WithPanicLevel() Option {
	return WithLevel(PanicLevel)
}

// Note.

// WithConsole enables related functionality.
func WithConsole(enabled bool) Option {
	return func(c *config) {
		c.console = enabled
	}
}

// WithNoColor disables related functionality.
func WithNoColor() Option {
	return func(c *config) {
		c.noColor = true
	}
}

// WithFile configures related behavior.
// Note.
func WithFile(path string) Option {
	return func(c *config) {
		c.filePath = path
	}
}

// WithFileRotation documents related behavior.
// maxSize documents related behavior.
// maxAge documents related behavior.
// maxBackups documents related behavior.
func WithFileRotation(maxSizeMB, maxAgeDays, maxBackups int) Option {
	return func(c *config) {
		c.maxSize = maxSizeMB
		c.maxAge = maxAgeDays
		c.maxBackups = maxBackups
	}
}

// WithLocalTime documents related behavior.
func WithLocalTime(enabled bool) Option {
	return func(c *config) {
		c.localTime = enabled
	}
}

// WithCompress enables related functionality.
func WithCompress(enabled bool) Option {
	return func(c *config) {
		c.compress = enabled
	}
}

// Note.

// WithTimeFormat configures related behavior.
// Note.
func WithTimeFormat(format string) Option {
	return func(c *config) {
		c.timeFormat = format
	}
}

// WithRFC3339Time documents related behavior.
func WithRFC3339Time() Option {
	return WithTimeFormat(time.RFC3339)
}

// WithISO8601Time documents related behavior.
func WithISO8601Time() Option {
	return WithTimeFormat("2006-01-02T15:04:05Z07:00")
}

// Note.

// WithGlobalLogger configures related behavior.
func WithGlobalLogger() Option {
	return func(c *config) {
		c.setGlobal = true
	}
}

// WithCaller enables related functionality.
func WithCaller(enabled bool) Option {
	return func(c *config) {
		c.addCaller = enabled
	}
}

// Note.

// DevelopmentConfig documents related behavior.
// Note.
// Note.
// Note.
func DevelopmentConfig() []Option {
	return []Option{
		WithConsole(true),
		WithDebugLevel(),
		WithCaller(true),
	}
}

// ProductionConfig documents related behavior.
// Note.
// Note.
// Note.
func ProductionConfig(logPath string) []Option {
	return []Option{
		WithConsole(false),
		WithInfoLevel(),
		WithFile(logPath),
		WithFileRotation(100, 7, 10),
		WithCompress(true),
	}
}
