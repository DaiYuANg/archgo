package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	collectionmapping "github.com/DaiYuANg/arcgo/collectionx/mapping"
	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/DaiYuANg/arcgo/httpx/adapter"
	"github.com/DaiYuANg/arcgo/httpx/adapter/fiber"
	"github.com/DaiYuANg/arcgo/logx"
	"github.com/DaiYuANg/arcgo/pkg/randomport"
	"github.com/danielgtaylor/huma/v2"
	"github.com/go-playground/validator/v10"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
)

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserStore struct {
	mu     sync.RWMutex
	nextID int
	users  *collectionmapping.Map[int, User]
}

func NewUserStore() *UserStore {
	now := time.Now().UTC()
	return &UserStore{
		nextID: 3,
		users: collectionmapping.NewMapFrom(map[int]User{
			1: {
				ID:        1,
				Name:      "Alice",
				Email:     "alice@example.com",
				Age:       26,
				CreatedAt: now,
				UpdatedAt: now,
			},
			2: {
				ID:        2,
				Name:      "Bob",
				Email:     "bob@example.com",
				Age:       30,
				CreatedAt: now,
				UpdatedAt: now,
			},
		}),
	}
}

func (s *UserStore) List(search string, limit, offset int) ([]User, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]User, 0, s.users.Len())
	s.users.Range(func(_ int, u User) bool {
		if search != "" && !strings.Contains(strings.ToLower(u.Name+u.Email), strings.ToLower(search)) {
			return true
		}
		items = append(items, u)
		return true
	})

	total := len(items)
	if offset >= total {
		return []User{}, total
	}

	end := min(offset+limit, total)
	return items[offset:end], total
}

func (s *UserStore) Get(id int) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.users.Get(id)
}

func (s *UserStore) Create(in CreateUserBody) User {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	u := User{
		ID:        s.nextID,
		Name:      in.Name,
		Email:     in.Email,
		Age:       in.Age,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.users.Set(u.ID, u)
	s.nextID++
	return u
}

func (s *UserStore) Update(id int, in UpdateUserBody) (User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.users.Get(id)
	if !ok {
		return User{}, false
	}

	if in.Name != nil {
		u.Name = *in.Name
	}
	if in.Email != nil {
		u.Email = *in.Email
	}
	if in.Age != nil {
		u.Age = *in.Age
	}
	u.UpdatedAt = time.Now().UTC()
	s.users.Set(id, u)
	return u, true
}

func (s *UserStore) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.users.Delete(id)
}

type ListUsersInput struct {
	Limit int    `query:"limit" validate:"omitempty,min=1,max=100"`
	Page  int    `query:"page" validate:"omitempty,min=1"`
	Q     string `query:"q" validate:"omitempty,max=100"`
}

type ListUsersOutput struct {
	Body struct {
		Items []User `json:"items"`
		Total int    `json:"total"`
		Page  int    `json:"page"`
		Limit int    `json:"limit"`
	} `json:"body"`
}

type GetUserInput struct {
	ID int `path:"id" validate:"required,min=1"`
}

type GetUserOutput struct {
	Body User `json:"body"`
}

type CreateUserBody struct {
	Name  string `json:"name" validate:"required,min=2,max=64"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"gte=0,lte=130"`
}

type CreateUserInput struct {
	Body CreateUserBody `json:"body"`
}

type CreateUserOutput struct {
	Body User `json:"body"`
}

type UpdateUserBody struct {
	Name  *string `json:"name,omitempty" validate:"omitempty,min=2,max=64"`
	Email *string `json:"email,omitempty" validate:"omitempty,email"`
	Age   *int    `json:"age,omitempty" validate:"omitempty,gte=0,lte=130"`
}

type UpdateUserInput struct {
	ID   int            `path:"id"`
	Body UpdateUserBody `json:"body"`
}

type UpdateUserOutput struct {
	Body User `json:"body"`
}

type DeleteUserInput struct {
	ID int `path:"id"`
}

type DeleteUserOutput struct {
	Body struct {
		Deleted bool `json:"deleted"`
	} `json:"body"`
}

type HealthOutput struct {
	Body struct {
		Status string `json:"status"`
		Time   string `json:"time"`
	} `json:"body"`
}

func main() {
	logger, err := logx.New(logx.WithConsole(true), logx.WithDebugLevel())
	if err != nil {
		panic(err)
	}
	defer func() { _ = logger.Close() }()

	store := NewUserStore()
	fiberAdapter := fiber.New(nil, adapter.HumaOptions{
		Title:       "ArcGo Fiber API",
		Version:     "1.0.0",
		Description: "Typed Fiber API example",
		DocsPath:    "/docs",
		OpenAPIPath: "/openapi.json",
	}).WithLogger(logx.NewSlog(logger))
	fiberAdapter.Router().Use(fiberrecover.New(), fiberlogger.New())

	server := httpx.NewServer(
		httpx.WithAdapter(fiberAdapter),
		httpx.WithLogger(logx.NewSlog(logger)),
		httpx.WithPrintRoutes(true),
		httpx.WithValidator(validator.New(validator.WithRequiredStructEnabled())),
	)

	httpx.MustGet(server, "/health", func(ctx context.Context, input *struct{}) (*HealthOutput, error) {
		out := &HealthOutput{}
		out.Body.Status = "ok"
		out.Body.Time = time.Now().UTC().Format(time.RFC3339)
		return out, nil
	}, huma.OperationTags("system"))

	api := server.Group("/api/v1")

	httpx.MustGroupGet(api, "/users", func(ctx context.Context, input *ListUsersInput) (*ListUsersOutput, error) {
		limit := input.Limit
		if limit <= 0 {
			limit = 10
		}
		if limit > 100 {
			limit = 100
		}

		page := input.Page
		if page <= 0 {
			page = 1
		}
		offset := (page - 1) * limit

		items, total := store.List(input.Q, limit, offset)
		out := &ListUsersOutput{}
		out.Body.Items = items
		out.Body.Total = total
		out.Body.Page = page
		out.Body.Limit = limit
		return out, nil
	}, huma.OperationTags("users"))

	httpx.MustGroupGet(api, "/users/{id}", func(ctx context.Context, input *GetUserInput) (*GetUserOutput, error) {
		u, ok := store.Get(input.ID)
		if !ok {
			return nil, httpx.NewError(404, "user not found")
		}
		out := &GetUserOutput{}
		out.Body = u
		return out, nil
	}, huma.OperationTags("users"))

	httpx.MustGroupPost(api, "/users", func(ctx context.Context, input *CreateUserInput) (*CreateUserOutput, error) {
		u := store.Create(input.Body)
		out := &CreateUserOutput{}
		out.Body = u
		return out, nil
	}, huma.OperationTags("users"))

	httpx.MustGroupPut(api, "/users/{id}", func(ctx context.Context, input *UpdateUserInput) (*UpdateUserOutput, error) {
		u, ok := store.Update(input.ID, input.Body)
		if !ok {
			return nil, httpx.NewError(404, "user not found")
		}
		out := &UpdateUserOutput{}
		out.Body = u
		return out, nil
	}, huma.OperationTags("users"))

	httpx.MustGroupDelete(api, "/users/{id}", func(ctx context.Context, input *DeleteUserInput) (*DeleteUserOutput, error) {
		deleted := store.Delete(input.ID)
		if !deleted {
			return nil, httpx.NewError(404, "user not found")
		}
		out := &DeleteUserOutput{}
		out.Body.Deleted = true
		return out, nil
	}, huma.OperationTags("users"))

	port := randomport.MustFind()
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Fiber example server running on %s\n", addr)
	fmt.Printf("GET  /health\n")
	fmt.Printf("GET  /api/v1/users?limit=10&page=1&q=alice\n")
	fmt.Printf("GET  /api/v1/users/{id}\n")
	fmt.Printf("POST /api/v1/users\n")
	fmt.Printf("PUT  /api/v1/users/{id}\n")
	fmt.Printf("DELETE /api/v1/users/{id}\n")
	fmt.Printf("OpenAPI: http://localhost%s/openapi.json\n", addr)
	fmt.Printf("Docs:    http://localhost%s/docs\n", addr)

	if err := server.ListenAndServe(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
