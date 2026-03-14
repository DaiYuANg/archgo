package main

import (
	"time"

	"github.com/uptrace/bun"
)

type userModel struct {
	bun.BaseModel `bun:"table:rbac_users,alias:u"`

	ID        int64     `bun:",pk,autoincrement"`
	Username  string    `bun:",notnull,unique"`
	Password  string    `bun:",notnull"`
	CreatedAt time.Time `bun:",notnull,default:current_timestamp"`
}

type roleModel struct {
	bun.BaseModel `bun:"table:rbac_roles,alias:r"`

	ID   int64  `bun:",pk,autoincrement"`
	Code string `bun:",notnull,unique"`
	Name string `bun:",notnull"`
}

type permissionModel struct {
	bun.BaseModel `bun:"table:rbac_permissions,alias:p"`

	ID       int64  `bun:",pk,autoincrement"`
	Action   string `bun:",notnull"`
	Resource string `bun:",notnull"`
}

type userRoleModel struct {
	bun.BaseModel `bun:"table:rbac_user_roles,alias:ur"`

	UserID int64 `bun:",pk"`
	RoleID int64 `bun:",pk"`
}

type rolePermissionModel struct {
	bun.BaseModel `bun:"table:rbac_role_permissions,alias:rp"`

	RoleID       int64 `bun:",pk"`
	PermissionID int64 `bun:",pk"`
}

type bookModel struct {
	bun.BaseModel `bun:"table:rbac_books,alias:b"`

	ID        int64     `bun:",pk,autoincrement"`
	Title     string    `bun:",notnull"`
	Author    string    `bun:",notnull"`
	CreatedBy int64     `bun:",notnull"`
	CreatedAt time.Time `bun:",notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:",notnull,default:current_timestamp"`
}

type appPrincipal struct {
	UserID   int64
	Username string
	Roles    []string
}

type loginInput struct {
	Body struct {
		Username string `json:"username" validate:"required,min=3,max=64"`
		Password string `json:"password" validate:"required,min=3,max=128"`
	} `json:"body"`
}

type loginOutput struct {
	Body struct {
		Token    string   `json:"token"`
		UserID   int64    `json:"user_id"`
		Username string   `json:"username"`
		Roles    []string `json:"roles"`
	} `json:"body"`
}

type listBooksOutput struct {
	Body struct {
		Items []bookItem `json:"items"`
		Total int        `json:"total"`
	} `json:"body"`
}

type bookItem struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	CreatedBy int64  `json:"created_by"`
}

type createBookInput struct {
	Body struct {
		Title  string `json:"title" validate:"required,min=1,max=200"`
		Author string `json:"author" validate:"required,min=1,max=120"`
	} `json:"body"`
}

type createBookOutput struct {
	Body bookItem `json:"body"`
}

type deleteBookInput struct {
	ID int64 `path:"id" validate:"required,min=1"`
}

type deleteBookOutput struct {
	Body struct {
		Deleted bool `json:"deleted"`
	} `json:"body"`
}
