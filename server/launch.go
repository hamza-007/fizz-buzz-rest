package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	config "fizz-buzz-rest/utils/config"

	errgroup "golang.org/x/sync/errgroup"
)

// TODO(lifetime): export rate, errors and duration as metrics; an autoscaler and
// an SLO both need a number, and today nothing reports saturation.
func Launch() error {
	logger := newLogger()

	server := http.Server{
		Addr:    config.APP().Addr(),
		Handler: routes(),

		ReadHeaderTimeout: config.APP().ReadHeaderTimeout,
		ReadTimeout:       config.APP().ReadTimeout,
		WriteTimeout:      config.APP().WriteTimeout,
		IdleTimeout:       config.APP().IdleTimeout,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		if err := Health(ctx); err != nil {
			return err
		}

		logger.Info("launch server",
			"addr", server.Addr,
			"env", config.Environment(),
			"max_limit", config.FizzBuzz().MaxLimit,
			"max_stats_keys", config.FizzBuzz().MaxStatsKeys,
			"rate_limit", config.APP().RateLimit,
			"rate_window", config.APP().RateWindow.String(),
			"cors_origins", config.APP().CORS,
			"trust_proxy", config.APP().TrustProxy,
		)

		err := server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	})

	group.Go(func() error {
		<-ctx.Done()

		logger.Info("shutdown server", "timeout", config.APP().ShutdownTimeout.String())

		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), config.APP().ShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()
			return err
		}
		return nil
	})

	if err := group.Wait(); err != nil {
		return err
	}

	logger.Info("server stopped cleanly")
	return nil
}

func newLogger() *slog.Logger {
	options := &slog.HandlerOptions{Level: config.APP().LogLevel}

	var handler slog.Handler = slog.NewJSONHandler(os.Stdout, options)
	if config.IsDevelopment() {
		handler = slog.NewTextHandler(os.Stdout, options)
	}

	logger := slog.New(handler)

	slog.SetDefault(logger)
	return logger
}
