package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/DaiYuANg/arcgo/eventx"
	"github.com/DaiYuANg/arcgo/logx"
	"github.com/DaiYuANg/arcgo/observabilityx"
	"go.uber.org/fx"
)

type loginSucceededEvent struct {
	UserID   int64
	Username string
	Roles    []string
}

func (e loginSucceededEvent) Name() string {
	return "auth.login.succeeded"
}

type bookCreatedEvent struct {
	BookID  int64
	Title   string
	Author  string
	ActorID int64
	Actor   string
}

func (e bookCreatedEvent) Name() string {
	return "rbac.book.created"
}

type bookDeletedEvent struct {
	BookID  int64
	ActorID int64
	Actor   string
}

func (e bookDeletedEvent) Name() string {
	return "rbac.book.deleted"
}

func registerEventSubscribers(
	lc fx.Lifecycle,
	bus eventx.BusRuntime,
	logger *slog.Logger,
	obs observabilityx.Observability,
) error {
	unsubscribers := make([]func(), 0, 3)

	loginUnsub, err := eventx.Subscribe(bus, func(ctx context.Context, event loginSucceededEvent) error {
		obs.AddCounter(ctx, "rbac_events_total", 1, observabilityx.String("event", event.Name()))
		logx.WithFields(logger, map[string]any{
			"event":    event.Name(),
			"user_id":  event.UserID,
			"username": event.Username,
			"roles":    event.Roles,
		}).Info("event handled")
		return nil
	})
	if err != nil {
		return fmt.Errorf("subscribe login event failed: %w", err)
	}
	unsubscribers = append(unsubscribers, loginUnsub)

	createdUnsub, err := eventx.Subscribe(bus, func(ctx context.Context, event bookCreatedEvent) error {
		obs.AddCounter(ctx, "rbac_events_total", 1, observabilityx.String("event", event.Name()))
		logx.WithFields(logger, map[string]any{
			"event":    event.Name(),
			"book_id":  event.BookID,
			"title":    event.Title,
			"actor_id": event.ActorID,
			"actor":    event.Actor,
		}).Info("event handled")
		return nil
	})
	if err != nil {
		return fmt.Errorf("subscribe book created event failed: %w", err)
	}
	unsubscribers = append(unsubscribers, createdUnsub)

	deletedUnsub, err := eventx.Subscribe(bus, func(ctx context.Context, event bookDeletedEvent) error {
		obs.AddCounter(ctx, "rbac_events_total", 1, observabilityx.String("event", event.Name()))
		logx.WithFields(logger, map[string]any{
			"event":    event.Name(),
			"book_id":  event.BookID,
			"actor_id": event.ActorID,
			"actor":    event.Actor,
		}).Info("event handled")
		return nil
	})
	if err != nil {
		return fmt.Errorf("subscribe book deleted event failed: %w", err)
	}
	unsubscribers = append(unsubscribers, deletedUnsub)

	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			for _, unsubscribe := range unsubscribers {
				if unsubscribe != nil {
					unsubscribe()
				}
			}
			return nil
		},
	})

	return nil
}
