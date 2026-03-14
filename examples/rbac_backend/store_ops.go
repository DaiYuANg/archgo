package main

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/samber/lo"
)

func (s *store) login(ctx context.Context, username string, password string) (appPrincipal, error) {
	var user userModel
	err := s.db.NewSelect().
		Model(&user).
		Where("username = ?", strings.TrimSpace(username)).
		Where("password = ?", password).
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appPrincipal{}, errors.New("invalid username or password")
		}
		return appPrincipal{}, err
	}

	roles, err := s.userRoles(ctx, user.ID)
	if err != nil {
		return appPrincipal{}, err
	}

	return appPrincipal{UserID: user.ID, Username: user.Username, Roles: roles}, nil
}

func (s *store) userRoles(ctx context.Context, userID int64) ([]string, error) {
	var roles []roleModel
	err := s.db.NewSelect().
		Model(&roles).
		Join("JOIN rbac_user_roles ur ON ur.role_id = r.id").
		Where("ur.user_id = ?", userID).
		OrderExpr("r.id ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return lo.Map(roles, func(item roleModel, _ int) string {
		return item.Code
	}), nil
}

func (s *store) can(ctx context.Context, userID int64, action string, resource string) (bool, error) {
	count, err := s.db.NewSelect().
		Model((*permissionModel)(nil)).
		Join("JOIN rbac_role_permissions rp ON rp.permission_id = p.id").
		Join("JOIN rbac_user_roles ur ON ur.role_id = rp.role_id").
		Where("ur.user_id = ?", userID).
		Where("p.action = ?", action).
		Where("p.resource = ?", resource).
		Count(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *store) listBooks(ctx context.Context) ([]bookItem, error) {
	var rows []bookModel
	if err := s.db.NewSelect().Model(&rows).OrderExpr("b.id ASC").Scan(ctx); err != nil {
		return nil, err
	}
	return lo.Map(rows, func(row bookModel, _ int) bookItem {
		return bookItem{ID: row.ID, Title: row.Title, Author: row.Author, CreatedBy: row.CreatedBy}
	}), nil
}

func (s *store) createBook(ctx context.Context, title string, author string, createdBy int64) (bookItem, error) {
	now := time.Now()
	row := bookModel{Title: title, Author: author, CreatedBy: createdBy, CreatedAt: now, UpdatedAt: now}
	if _, err := s.db.NewInsert().Model(&row).Exec(ctx); err != nil {
		return bookItem{}, err
	}
	return bookItem{ID: row.ID, Title: row.Title, Author: row.Author, CreatedBy: row.CreatedBy}, nil
}

func (s *store) deleteBook(ctx context.Context, id int64) (bool, error) {
	res, err := s.db.NewDelete().Model((*bookModel)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}
