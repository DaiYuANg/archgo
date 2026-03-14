package bunx

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/samber/lo"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/driver/sqliteshim"

	_ "github.com/go-sql-driver/mysql"
)

// Driver constants supported by bunx.Open.
const (
	DriverSQLite   = "sqlite"
	DriverMySQL    = "mysql"
	DriverPostgres = "postgres"
)

type dbOptions struct {
	logger        *slog.Logger
	slowThreshold time.Duration
	logQuery      bool
	logArgs       bool
}

// DBOption customizes bunx DB bootstrap behavior.
type DBOption func(*dbOptions)

// WithLogger injects the slog logger used for bun query logs.
func WithLogger(logger *slog.Logger) DBOption {
	return func(opts *dbOptions) {
		if logger != nil {
			opts.logger = logger
		}
	}
}

// WithSlowQueryThreshold marks queries slower than threshold as warn logs.
func WithSlowQueryThreshold(threshold time.Duration) DBOption {
	return func(opts *dbOptions) {
		opts.slowThreshold = threshold
	}
}

// WithQueryLogging enables or disables query logging.
func WithQueryLogging(enabled bool) DBOption {
	return func(opts *dbOptions) {
		opts.logQuery = enabled
	}
}

// WithQueryArgs enables or disables query args logging.
func WithQueryArgs(enabled bool) DBOption {
	return func(opts *dbOptions) {
		opts.logArgs = enabled
	}
}

// Open creates a bun DB for the given driver and DSN.
func Open(driver string, dsn string, opts ...DBOption) (*bun.DB, error) {
	var (
		sqlDB *sql.DB
		err   error
		db    *bun.DB
	)

	switch driver {
	case DriverSQLite:
		sqlDB, err = sql.Open(sqliteshim.ShimName, dsn)
		if err != nil {
			return nil, fmt.Errorf("open sqlite failed: %w", err)
		}
		db = bun.NewDB(sqlDB, sqlitedialect.New())
	case DriverMySQL:
		sqlDB, err = sql.Open(DriverMySQL, dsn)
		if err != nil {
			return nil, fmt.Errorf("open mysql failed: %w", err)
		}
		db = bun.NewDB(sqlDB, mysqldialect.New())
	case DriverPostgres:
		sqlDB = sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
		db = bun.NewDB(sqlDB, pgdialect.New())
	default:
		return nil, fmt.Errorf("unsupported db driver: %s", driver)
	}

	return Wrap(db, opts...), nil
}

// Wrap adds bunx defaults (notably slog query logging) to an existing bun DB.
func Wrap(db *bun.DB, opts ...DBOption) *bun.DB {
	if db == nil {
		return nil
	}

	cfg := dbOptions{
		logger:        slog.Default(),
		slowThreshold: 300 * time.Millisecond,
		logQuery:      true,
		logArgs:       false,
	}
	lo.ForEach(opts, func(opt DBOption, _ int) {
		if opt != nil {
			opt(&cfg)
		}
	})

	if cfg.logQuery && cfg.logger != nil {
		db.AddQueryHook(newQueryLogHook(cfg.logger, cfg.slowThreshold, cfg.logQuery, cfg.logArgs))
	}

	return db
}
