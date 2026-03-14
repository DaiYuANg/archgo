package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func main() {
	app := fx.New(
		newAppModule(),
		fx.WithLogger(func(logger *slog.Logger) fxevent.Logger {
			return &fxevent.SlogLogger{Logger: logger}
		}),
	)

	startCtx, startCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer startCancel()
	if err := app.Start(startCtx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "start app failed: %v\n", err)
		os.Exit(1)
	}

	<-app.Done()

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer stopCancel()
	if err := app.Stop(stopCtx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "stop app failed: %v\n", err)
		os.Exit(1)
	}
}
