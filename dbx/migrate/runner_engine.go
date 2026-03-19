package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/DaiYuANg/arcgo/collectionx"
	"github.com/pressly/goose/v3"
)

type runnerEngine struct {
	runner        *Runner
	engine        *goose.Provider
	metaByVersion collectionx.Map[int64, AppliedRecord]
}

func (r *Runner) newRunnerEngineForGo(migrations []Migration) (*runnerEngine, error) {
	if len(migrations) == 0 {
		return nil, nil
	}

	gooseMigrations := collectionx.NewListWithCapacity[*goose.Migration](len(migrations))
	metaByVersion := collectionx.NewMapWithCapacity[int64, AppliedRecord](len(migrations))
	for _, migration := range migrations {
		version, err := parseNumericVersion(migration.Version())
		if err != nil {
			return nil, err
		}

		gooseMigrations.Add(goose.NewGoMigration(
			version,
			&goose.GoFunc{RunTx: migration.Up},
			&goose.GoFunc{RunTx: migration.Down},
		))
		metaByVersion.Set(version, AppliedRecord{
			Version:     migration.Version(),
			Description: migration.Description(),
			Kind:        KindGo,
			Checksum:    checksumGoMigration(migration),
			Success:     true,
		})
	}

	return r.newRunnerEngine(gooseMigrations.Values(), metaByVersion)
}

func (r *Runner) newRunnerEngineForSQL(source FileSource) (*runnerEngine, []loadedSQLMigration, error) {
	loaded, err := loadSQLMigrations(source)
	if err != nil {
		return nil, nil, err
	}
	if len(loaded) == 0 {
		return nil, nil, nil
	}

	gooseMigrations := collectionx.NewListWithCapacity[*goose.Migration](len(loaded))
	metaByVersion := collectionx.NewMapWithCapacity[int64, AppliedRecord](len(loaded))
	repeatables := collectionx.NewList[loadedSQLMigration]()
	for _, migration := range loaded {
		if migration.kind == KindRepeatable {
			repeatables.Add(migration)
			continue
		}

		version, err := parseNumericVersion(migration.Version)
		if err != nil {
			return nil, nil, err
		}

		gooseMigrations.Add(goose.NewGoMigration(
			version,
			runTxSQL(migration.upSQL),
			runTxSQL(migration.downSQL),
		))
		metaByVersion.Set(version, AppliedRecord{
			Version:     migration.Version,
			Description: migration.Description,
			Kind:        migration.kind,
			Checksum:    migration.checksum,
			Success:     true,
		})
	}

	if gooseMigrations.Len() == 0 {
		return nil, repeatables.Values(), nil
	}
	engine, err := r.newRunnerEngine(gooseMigrations.Values(), metaByVersion)
	if err != nil {
		return nil, nil, err
	}
	return engine, repeatables.Values(), nil
}

func (r *Runner) newRunnerEngine(migrations []*goose.Migration, metaByVersion collectionx.Map[int64, AppliedRecord]) (*runnerEngine, error) {
	if len(migrations) == 0 {
		return nil, nil
	}

	engine, err := goose.NewProvider(
		goose.DialectCustom,
		r.db,
		nil,
		goose.WithStore(newHistoryStore(r.dialect, r.options.HistoryTable, metaByVersion)),
		goose.WithDisableGlobalRegistry(true),
		goose.WithAllowOutofOrder(r.options.AllowOutOfOrder),
		goose.WithGoMigrations(migrations...),
	)
	if err != nil {
		return nil, err
	}
	return &runnerEngine{
		runner:        r,
		engine:        engine,
		metaByVersion: metaByVersion,
	}, nil
}

func runTxSQL(statement string) *goose.GoFunc {
	if statement == "" {
		return nil
	}
	return &goose.GoFunc{
		RunTx: func(ctx context.Context, tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx, statement)
			return err
		},
	}
}

func parseNumericVersion(version string) (int64, error) {
	parsed, err := strconv.ParseInt(version, 10, 64)
	if err != nil || parsed < 1 {
		return 0, fmt.Errorf("dbx/migrate: goose requires a positive numeric version, got %q", version)
	}
	return parsed, nil
}

