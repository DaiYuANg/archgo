package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/DaiYuANg/arcgo/observabilityx"
	"github.com/samber/lo"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/driver/sqliteshim"
	"go.uber.org/fx"

	_ "github.com/go-sql-driver/mysql"
)

type store struct {
	db     *bun.DB
	obs    observabilityx.Observability
	logger *slog.Logger
}

func newStore(
	lc fx.Lifecycle,
	cfg appConfig,
	obs observabilityx.Observability,
	logger *slog.Logger,
) (*store, error) {
	db, err := openBunDB(cfg)
	if err != nil {
		return nil, err
	}

	s := &store{
		db:     db,
		obs:    obs,
		logger: logger,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := s.initSchema(ctx); err != nil {
				return err
			}
			if err := s.seed(ctx); err != nil {
				return err
			}
			return nil
		},
		OnStop: func(context.Context) error {
			return s.close()
		},
	})

	return s, nil
}

func openBunDB(cfg appConfig) (*bun.DB, error) {
	switch cfg.dbDriver() {
	case "sqlite":
		sqlDB, err := sql.Open(sqliteshim.ShimName, cfg.dbDSN())
		if err != nil {
			return nil, fmt.Errorf("open sqlite failed: %w", err)
		}
		return bun.NewDB(sqlDB, sqlitedialect.New()), nil
	case "mysql":
		sqlDB, err := sql.Open("mysql", cfg.dbDSN())
		if err != nil {
			return nil, fmt.Errorf("open mysql failed: %w", err)
		}
		return bun.NewDB(sqlDB, mysqldialect.New()), nil
	case "postgres":
		sqlDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(cfg.dbDSN())))
		return bun.NewDB(sqlDB, pgdialect.New()), nil
	default:
		return nil, fmt.Errorf("unsupported db driver: %s", cfg.dbDriver())
	}
}

func (s *store) close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *store) initSchema(ctx context.Context) error {
	ctx, span := s.obs.StartSpan(ctx, "rbac.store.init_schema")
	defer span.End()

	models := []any{
		(*userModel)(nil),
		(*roleModel)(nil),
		(*permissionModel)(nil),
		(*userRoleModel)(nil),
		(*rolePermissionModel)(nil),
		(*bookModel)(nil),
	}
	for _, model := range models {
		if _, err := s.db.NewCreateTable().Model(model).IfNotExists().Exec(ctx); err != nil {
			span.RecordError(err)
			return err
		}
	}
	return nil
}

func (s *store) seed(ctx context.Context) error {
	ctx, span := s.obs.StartSpan(ctx, "rbac.store.seed")
	defer span.End()

	count, err := s.db.NewSelect().Model((*userModel)(nil)).Count(ctx)
	if err != nil {
		span.RecordError(err)
		return err
	}
	if count > 0 {
		return nil
	}

	roles := []roleModel{{Code: "admin", Name: "Administrator"}, {Code: "user", Name: "User"}}
	if _, err = s.db.NewInsert().Model(&roles).Exec(ctx); err != nil {
		span.RecordError(err)
		return err
	}

	var roleRows []roleModel
	if err = s.db.NewSelect().Model(&roleRows).Scan(ctx); err != nil {
		span.RecordError(err)
		return err
	}
	roleIDs := lo.SliceToMap(roleRows, func(item roleModel) (string, int64) {
		return item.Code, item.ID
	})

	permissions := []permissionModel{
		{Action: "query", Resource: "book"},
		{Action: "create", Resource: "book"},
		{Action: "delete", Resource: "book"},
	}
	if _, err = s.db.NewInsert().Model(&permissions).Exec(ctx); err != nil {
		span.RecordError(err)
		return err
	}

	var permissionRows []permissionModel
	if err = s.db.NewSelect().Model(&permissionRows).Scan(ctx); err != nil {
		span.RecordError(err)
		return err
	}
	permissionIDs := lo.SliceToMap(permissionRows, func(item permissionModel) (string, int64) {
		return item.Action + ":" + item.Resource, item.ID
	})

	rolePermissions := []rolePermissionModel{
		{RoleID: roleIDs["admin"], PermissionID: permissionIDs["query:book"]},
		{RoleID: roleIDs["admin"], PermissionID: permissionIDs["create:book"]},
		{RoleID: roleIDs["admin"], PermissionID: permissionIDs["delete:book"]},
		{RoleID: roleIDs["user"], PermissionID: permissionIDs["query:book"]},
	}
	if _, err = s.db.NewInsert().Model(&rolePermissions).Exec(ctx); err != nil {
		span.RecordError(err)
		return err
	}

	users := []userModel{
		{Username: "alice", Password: "admin123"},
		{Username: "bob", Password: "user123"},
	}
	if _, err = s.db.NewInsert().Model(&users).Exec(ctx); err != nil {
		span.RecordError(err)
		return err
	}
	if err = s.db.NewSelect().Model(&users).Scan(ctx); err != nil {
		span.RecordError(err)
		return err
	}

	userRoles := []userRoleModel{
		{UserID: users[0].ID, RoleID: roleIDs["admin"]},
		{UserID: users[1].ID, RoleID: roleIDs["user"]},
	}
	if _, err = s.db.NewInsert().Model(&userRoles).Exec(ctx); err != nil {
		span.RecordError(err)
		return err
	}

	books := []bookModel{
		{Title: "Distributed Systems", Author: "Tanenbaum", CreatedBy: users[0].ID},
		{Title: "Go in Action", Author: "Kennedy", CreatedBy: users[0].ID},
	}
	if _, err = s.db.NewInsert().Model(&books).Exec(ctx); err != nil {
		span.RecordError(err)
		return err
	}

	s.logger.Info("seed data initialized")
	return nil
}
