package core

import (
	"context"
	"log/slog"

	"github.com/DaiYuANg/arcgo/bunx"
	"github.com/DaiYuANg/arcgo/examples/rbac_backend/internal/config"
	"github.com/DaiYuANg/arcgo/examples/rbac_backend/internal/entity"
	"github.com/DaiYuANg/arcgo/observabilityx"
	"github.com/samber/lo"
	"github.com/uptrace/bun"
	"go.uber.org/fx"
)

type Store struct {
	db     *bun.DB
	obs    observabilityx.Observability
	logger *slog.Logger
	ownDB  bool
}

type storeDeps struct {
	fx.In

	Lifecycle  fx.Lifecycle
	Config     config.AppConfig
	Obs        observabilityx.Observability
	Logger     *slog.Logger
	ExternalDB *bun.DB `optional:"true"`
}

func NewStore(deps storeDeps) (*Store, error) {
	var (
		db    *bun.DB
		err   error
		ownDB bool
	)

	if deps.ExternalDB != nil {
		db = deps.ExternalDB
		ownDB = false
	} else {
		db, err = openBunDB(deps.Config, deps.Logger)
		if err != nil {
			return nil, err
		}
		ownDB = true
	}

	s := &Store{
		db:     db,
		obs:    deps.Obs,
		logger: deps.Logger,
		ownDB:  ownDB,
	}

	deps.Lifecycle.Append(fx.Hook{
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

func (s *Store) DB() *bun.DB {
	if s == nil {
		return nil
	}
	return s.db
}

func (s *Store) Logger() *slog.Logger {
	if s == nil {
		return nil
	}
	return s.logger
}

func openBunDB(cfg config.AppConfig, logger *slog.Logger) (*bun.DB, error) {
	db, err := bunx.Open(
		cfg.DBDriver(),
		cfg.DBDSN(),
		bunx.WithLogger(logger),
	)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (s *Store) close() error {
	if s == nil || s.db == nil || !s.ownDB {
		return nil
	}
	return s.db.Close()
}

func (s *Store) initSchema(ctx context.Context) error {
	ctx, span := s.obs.StartSpan(ctx, "rbac.store.init_schema")
	defer span.End()

	models := []any{
		(*entity.UserModel)(nil),
		(*entity.RoleModel)(nil),
		(*entity.PermissionModel)(nil),
		(*entity.UserRoleModel)(nil),
		(*entity.RolePermissionModel)(nil),
		(*entity.BookModel)(nil),
	}
	err := lo.Reduce(models, func(acc error, model any, _ int) error {
		if acc != nil {
			return acc
		}
		if _, createErr := s.db.NewCreateTable().Model(model).IfNotExists().Exec(ctx); createErr != nil {
			span.RecordError(createErr)
			return createErr
		}
		return nil
	}, nil)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) seed(ctx context.Context) error {
	ctx, span := s.obs.StartSpan(ctx, "rbac.store.seed")
	defer span.End()

	count, err := s.db.NewSelect().Model((*entity.UserModel)(nil)).Count(ctx)
	if err != nil {
		span.RecordError(err)
		return err
	}
	if count > 0 {
		return nil
	}

	roles := []entity.RoleModel{{Code: "admin", Name: "Administrator"}, {Code: "user", Name: "User"}}
	if _, err = s.db.NewInsert().Model(&roles).Exec(ctx); err != nil {
		span.RecordError(err)
		return err
	}

	var roleRows []entity.RoleModel
	if err = s.db.NewSelect().Model(&roleRows).Scan(ctx); err != nil {
		span.RecordError(err)
		return err
	}
	roleIDs := lo.SliceToMap(roleRows, func(item entity.RoleModel) (string, int64) {
		return item.Code, item.ID
	})

	adminResources := []string{"book", "user", "role"}
	adminActions := []string{"query", "create", "update", "delete"}
	permissions := lo.FlatMap(adminResources, func(resource string, _ int) []entity.PermissionModel {
		return lo.Map(adminActions, func(action string, _ int) entity.PermissionModel {
			return entity.PermissionModel{
				Action:   action,
				Resource: resource,
			}
		})
	})
	if _, err = s.db.NewInsert().Model(&permissions).Exec(ctx); err != nil {
		span.RecordError(err)
		return err
	}

	var permissionRows []entity.PermissionModel
	if err = s.db.NewSelect().Model(&permissionRows).Scan(ctx); err != nil {
		span.RecordError(err)
		return err
	}
	permissionIDs := lo.SliceToMap(permissionRows, func(item entity.PermissionModel) (string, int64) {
		return item.Action + ":" + item.Resource, item.ID
	})

	rolePermissions := []entity.RolePermissionModel{
		{RoleID: roleIDs["admin"], PermissionID: permissionIDs["query:book"]},
		{RoleID: roleIDs["admin"], PermissionID: permissionIDs["create:book"]},
		{RoleID: roleIDs["admin"], PermissionID: permissionIDs["update:book"]},
		{RoleID: roleIDs["admin"], PermissionID: permissionIDs["delete:book"]},
		{RoleID: roleIDs["admin"], PermissionID: permissionIDs["query:user"]},
		{RoleID: roleIDs["admin"], PermissionID: permissionIDs["create:user"]},
		{RoleID: roleIDs["admin"], PermissionID: permissionIDs["update:user"]},
		{RoleID: roleIDs["admin"], PermissionID: permissionIDs["delete:user"]},
		{RoleID: roleIDs["admin"], PermissionID: permissionIDs["query:role"]},
		{RoleID: roleIDs["admin"], PermissionID: permissionIDs["create:role"]},
		{RoleID: roleIDs["admin"], PermissionID: permissionIDs["update:role"]},
		{RoleID: roleIDs["admin"], PermissionID: permissionIDs["delete:role"]},
		{RoleID: roleIDs["user"], PermissionID: permissionIDs["query:book"]},
	}
	if _, err = s.db.NewInsert().Model(&rolePermissions).Exec(ctx); err != nil {
		span.RecordError(err)
		return err
	}

	users := []entity.UserModel{
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

	userRoles := []entity.UserRoleModel{
		{UserID: users[0].ID, RoleID: roleIDs["admin"]},
		{UserID: users[1].ID, RoleID: roleIDs["user"]},
	}
	if _, err = s.db.NewInsert().Model(&userRoles).Exec(ctx); err != nil {
		span.RecordError(err)
		return err
	}

	books := []entity.BookModel{
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
