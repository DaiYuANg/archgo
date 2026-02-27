package logx

type Option func(*config)

type config struct {
	level      string
	console    bool
	filePath   string
	maxSize    int
	maxAge     int
	maxBackups int
	timeFormat string
	setGlobal  bool
}

func defaultConfig() config {
	return config{
		level:      "info",
		console:    true,
		timeFormat: "2006-01-02 15:04:05",
		maxSize:    100,
		maxAge:     7,
		maxBackups: 10,
	}
}

func WithLevel(level string) Option {
	return func(c *config) {
		c.level = level
	}
}

func WithConsole(enabled bool) Option {
	return func(c *config) {
		c.console = enabled
	}
}

func WithFile(path string) Option {
	return func(c *config) {
		c.filePath = path
	}
}

func WithFileRotation(maxSizeMB, maxAgeDays, maxBackups int) Option {
	return func(c *config) {
		c.maxSize = maxSizeMB
		c.maxAge = maxAgeDays
		c.maxBackups = maxBackups
	}
}

func WithTimeFormat(format string) Option {
	return func(c *config) {
		c.timeFormat = format
	}
}

func WithGlobalLogger() Option {
	return func(c *config) {
		c.setGlobal = true
	}
}
