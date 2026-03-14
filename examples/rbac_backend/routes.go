package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/DaiYuANg/arcgo/authx"
	"github.com/DaiYuANg/arcgo/eventx"
	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/DaiYuANg/arcgo/logx"
	"github.com/DaiYuANg/arcgo/observabilityx"
)

func registerHTTPRoutes(
	server httpx.ServerRuntime,
	s *store,
	jwtSvc *jwtService,
	bus eventx.BusRuntime,
	obs observabilityx.Observability,
	logger *slog.Logger,
) {
	httpx.MustPost(server, "/login", func(ctx context.Context, input *loginInput) (*loginOutput, error) {
		ctx, span, done := beginOperation(ctx, obs, "rbac.route.login")
		defer done()
		defer span.End()

		principal, err := s.login(ctx, input.Body.Username, input.Body.Password)
		if err != nil {
			span.RecordError(err)
			obs.AddCounter(ctx, "rbac_route_total", 1,
				observabilityx.String("route", "login"),
				observabilityx.String("result", "denied"),
			)
			return nil, httpx.NewError(401, "invalid username or password")
		}

		token, err := jwtSvc.issueToken(principal)
		if err != nil {
			span.RecordError(err)
			return nil, httpx.NewError(500, "issue jwt failed")
		}

		publishAsync(ctx, bus, loginSucceededEvent{
			UserID:   principal.UserID,
			Username: principal.Username,
			Roles:    principal.Roles,
		}, logger)

		obs.AddCounter(ctx, "rbac_route_total", 1,
			observabilityx.String("route", "login"),
			observabilityx.String("result", "ok"),
		)

		out := &loginOutput{}
		out.Body.Token = token
		out.Body.UserID = principal.UserID
		out.Body.Username = principal.Username
		out.Body.Roles = principal.Roles
		return out, nil
	})

	httpx.MustGet(server, "/books", func(ctx context.Context, _ *struct{}) (*listBooksOutput, error) {
		ctx, span, done := beginOperation(ctx, obs, "rbac.route.list_books")
		defer done()
		defer span.End()

		items, err := s.listBooks(ctx)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}

		obs.AddCounter(ctx, "rbac_route_total", 1,
			observabilityx.String("route", "list_books"),
			observabilityx.String("result", "ok"),
		)
		out := &listBooksOutput{}
		out.Body.Items = items
		out.Body.Total = len(items)
		return out, nil
	})

	httpx.MustPost(server, "/books", func(ctx context.Context, input *createBookInput) (*createBookOutput, error) {
		ctx, span, done := beginOperation(ctx, obs, "rbac.route.create_book")
		defer done()
		defer span.End()

		principal, ok := authx.PrincipalFromContextAs[appPrincipal](ctx)
		if !ok {
			return nil, httpx.NewError(401, "principal not found")
		}

		item, err := s.createBook(ctx, input.Body.Title, input.Body.Author, principal.UserID)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}

		publishAsync(ctx, bus, bookCreatedEvent{
			BookID:  item.ID,
			Title:   item.Title,
			Author:  item.Author,
			ActorID: principal.UserID,
			Actor:   principal.Username,
		}, logger)

		obs.AddCounter(ctx, "rbac_route_total", 1,
			observabilityx.String("route", "create_book"),
			observabilityx.String("result", "ok"),
		)
		out := &createBookOutput{}
		out.Body = item
		return out, nil
	})

	httpx.MustDelete(server, "/books/{id}", func(ctx context.Context, input *deleteBookInput) (*deleteBookOutput, error) {
		ctx, span, done := beginOperation(ctx, obs, "rbac.route.delete_book")
		defer done()
		defer span.End()

		principal, ok := authx.PrincipalFromContextAs[appPrincipal](ctx)
		if !ok {
			return nil, httpx.NewError(401, "principal not found")
		}

		deleted, err := s.deleteBook(ctx, input.ID)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}
		if !deleted {
			return nil, httpx.NewError(404, "book not found")
		}

		publishAsync(ctx, bus, bookDeletedEvent{
			BookID:  input.ID,
			ActorID: principal.UserID,
			Actor:   principal.Username,
		}, logger)

		obs.AddCounter(ctx, "rbac_route_total", 1,
			observabilityx.String("route", "delete_book"),
			observabilityx.String("result", "ok"),
		)
		out := &deleteBookOutput{}
		out.Body.Deleted = true
		return out, nil
	})
}

func publishAsync(ctx context.Context, bus eventx.BusRuntime, event eventx.Event, logger *slog.Logger) {
	if event == nil {
		return
	}
	if err := bus.PublishAsync(ctx, event); err != nil {
		logx.WithError(logx.WithFields(logger, map[string]any{"event": event.Name()}), err).
			Warn("publish async event failed")
	}
}

func beginOperation(
	ctx context.Context,
	obs observabilityx.Observability,
	name string,
) (context.Context, observabilityx.Span, func()) {
	started := time.Now()
	ctx, span := obs.StartSpan(ctx, name)
	return ctx, span, func() {
		obs.RecordHistogram(ctx, "rbac_route_duration_ms", float64(time.Since(started).Milliseconds()),
			observabilityx.String("operation", name),
		)
	}
}
