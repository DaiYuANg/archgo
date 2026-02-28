package logx

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Option 日志配置选项函数
type Option func(*config)

// config 日志配置
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

// defaultConfig 默认配置
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

// validate 验证配置
func (c *config) validate() error {
	// 验证日志级别
	if c.level < TraceLevel || c.level > DisabledLevel {
		return fmt.Errorf("invalid log level: %v", c.level)
	}

	// 验证文件配置
	if c.filePath != "" {
		// 检查文件目录是否可写
		dir := filepath.Dir(c.filePath)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			// 目录不存在，尝试创建
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("cannot create log directory %s: %w", dir, err)
			}
		}
	}

	// 验证文件大小限制
	if c.maxSize < 1 {
		return fmt.Errorf("maxSize must be at least 1MB, got %d", c.maxSize)
	}

	// 验证保留天数
	if c.maxAge < 0 {
		return fmt.Errorf("maxAge cannot be negative, got %d", c.maxAge)
	}

	// 验证备份数量
	if c.maxBackups < 0 {
		return fmt.Errorf("maxBackups cannot be negative, got %d", c.maxBackups)
	}

	return nil
}

// 日志级别选项

// WithLevel 设置日志级别
// 使用强类型 Level
func WithLevel(level Level) Option {
	return func(c *config) {
		c.level = level
	}
}

// WithLevelString 从字符串设置日志级别
func WithLevelString(level string) Option {
	return func(c *config) {
		l, err := ParseLevel(level)
		if err == nil {
			c.level = l
		}
	}
}

// WithTraceLevel 启用 trace 级别日志（包含所有日志）
func WithTraceLevel() Option {
	return WithLevel(TraceLevel)
}

// WithDebugLevel 启用 debug 级别日志
func WithDebugLevel() Option {
	return WithLevel(DebugLevel)
}

// WithInfoLevel 启用 info 级别日志
func WithInfoLevel() Option {
	return WithLevel(InfoLevel)
}

// WithWarnLevel 启用 warn 级别日志
func WithWarnLevel() Option {
	return WithLevel(WarnLevel)
}

// WithErrorLevel 启用 error 级别日志
func WithErrorLevel() Option {
	return WithLevel(ErrorLevel)
}

// WithFatalLevel 启用 fatal 级别日志
func WithFatalLevel() Option {
	return WithLevel(FatalLevel)
}

// WithPanicLevel 启用 panic 级别日志
func WithPanicLevel() Option {
	return WithLevel(PanicLevel)
}

// 输出选项

// WithConsole 启用/禁用控制台输出
func WithConsole(enabled bool) Option {
	return func(c *config) {
		c.console = enabled
	}
}

// WithNoColor 禁用控制台颜色输出
func WithNoColor() Option {
	return func(c *config) {
		c.noColor = true
	}
}

// WithFile 设置日志文件路径
// 路径目录会自动创建
func WithFile(path string) Option {
	return func(c *config) {
		c.filePath = path
	}
}

// WithFileRotation 配置文件轮转参数
// maxSize: 单个文件最大大小 (MB)
// maxAge: 文件保留天数
// maxBackups: 最大备份文件数量
func WithFileRotation(maxSizeMB, maxAgeDays, maxBackups int) Option {
	return func(c *config) {
		c.maxSize = maxSizeMB
		c.maxAge = maxAgeDays
		c.maxBackups = maxBackups
	}
}

// WithLocalTime 使用本地时间进行文件轮转
func WithLocalTime(enabled bool) Option {
	return func(c *config) {
		c.localTime = enabled
	}
}

// WithCompress 启用/禁用压缩备份文件
func WithCompress(enabled bool) Option {
	return func(c *config) {
		c.compress = enabled
	}
}

// 格式选项

// WithTimeFormat 设置时间格式
// 使用 Go 的时间格式：2006-01-02 15:04:05
func WithTimeFormat(format string) Option {
	return func(c *config) {
		c.timeFormat = format
	}
}

// WithRFC3339Time 使用 RFC3339 时间格式
func WithRFC3339Time() Option {
	return WithTimeFormat(time.RFC3339)
}

// WithISO8601Time 使用 ISO8601 时间格式
func WithISO8601Time() Option {
	return WithTimeFormat("2006-01-02T15:04:05Z07:00")
}

// 全局选项

// WithGlobalLogger 设置为全局默认 logger
func WithGlobalLogger() Option {
	return func(c *config) {
		c.setGlobal = true
	}
}

// WithCaller 启用/禁用调用者信息
func WithCaller(enabled bool) Option {
	return func(c *config) {
		c.addCaller = enabled
	}
}

// 预设配置

// DevelopmentConfig 开发环境配置
// - 启用控制台输出
// - Debug 级别
// - 启用调用者信息
func DevelopmentConfig() []Option {
	return []Option{
		WithConsole(true),
		WithDebugLevel(),
		WithCaller(true),
	}
}

// ProductionConfig 生产环境配置
// - 禁用控制台输出
// - Info 级别
// - 启用文件轮转
func ProductionConfig(logPath string) []Option {
	return []Option{
		WithConsole(false),
		WithInfoLevel(),
		WithFile(logPath),
		WithFileRotation(100, 7, 10),
		WithCompress(true),
	}
}
