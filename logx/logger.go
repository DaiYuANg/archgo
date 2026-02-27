package logx

import (
	"io"
	"os"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/samber/lo"
	oopszerolog "github.com/samber/oops/loggers/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	zerolog.Logger
	closers []io.Closer
}

func (l *Logger) Close() error {
	lo.ForEach(l.closers, func(closer io.Closer, index int) {
		if closer != nil {
			_ = closer.Close()
		}
	})
	return nil
}

func New(opts ...Option) (*Logger, error) {
	cfg := defaultConfig()

	for _, opt := range opts {
		opt(&cfg)
	}

	var writers []io.Writer
	var closers []io.Closer

	// console
	if cfg.console {
		writers = append(writers, zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: cfg.timeFormat,
		})
	}

	// file
	if cfg.filePath != "" {
		lj := &lumberjack.Logger{
			Filename:   cfg.filePath,
			MaxSize:    cfg.maxSize,
			MaxAge:     cfg.maxAge,
			MaxBackups: cfg.maxBackups,
		}
		writers = append(writers, lj)
		closers = append(closers, lj)
	}

	if len(writers) == 0 {
		writers = append(writers, os.Stdout)
	}

	level, err := zerolog.ParseLevel(cfg.level)
	if err != nil {
		level = zerolog.InfoLevel
	}

	mw := io.MultiWriter(writers...)

	z := zerolog.New(mw).
		Level(level).
		With().
		Timestamp().
		Logger()

	// better stack trace
	zerolog.ErrorStackMarshaler = oopszerolog.OopsStackMarshaller
	zerolog.ErrorMarshalFunc = oopszerolog.OopsMarshalFunc

	if cfg.setGlobal {
		zlog.Logger = z
	}

	return &Logger{
		Logger:  z,
		closers: closers,
	}, nil
}
