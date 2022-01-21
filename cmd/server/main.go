package main

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/apex/log"
	"github.com/coreos/go-systemd/daemon"
	"github.com/gorilla/handlers"
	"github.com/sewiti/licensing-system/internal/db"
	"github.com/sewiti/licensing-system/internal/web"
	"github.com/vrischmann/envconfig"
)

func main() {
	// Config
	var cfg config
	err := envconfig.Init(&cfg)
	if err != nil {
		log.WithError(err).Fatal("initializing environment configuration")
		return
	}
	if len(cfg.HTTP.CORS.AllowedOrigins) == 0 {
		log.Warn("HTTP_CORS_ALLOWED_ORIGINS not set")
	}

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Database
	err = db.MigrateUp(cfg.DB.DataSource)
	if err != nil {
		log.WithError(err).Fatal("migrating database")
		return
	}
	dbh, err := db.Open(cfg.DB.DataSource)
	if err != nil {
		log.WithError(err).Fatal("opening database")
		return
	}

	// HTTP Handler
	rt := web.NewRuntime(dbh)
	r := web.NewRouter(rt)
	h := handlers.CORS(
		handlers.AllowedHeaders(cfg.HTTP.CORS.AllowedHeaders),
		handlers.AllowedMethods(cfg.HTTP.CORS.AllowedMethods),
		handlers.AllowedOrigins(cfg.HTTP.CORS.AllowedOrigins),
	)(r)

	// HTTP Server
	srv := http.Server{
		Addr:         cfg.HTTP.Listen,
		Handler:      h,
		ReadTimeout:  cfg.HTTP.Timeout,
		WriteTimeout: cfg.HTTP.Timeout,
	}
	go func() {
		defer stop()
		err := srv.ListenAndServe()
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			log.WithError(err).Error("listening and serving http")
		}
	}()

	// Systemd notify
	daemon.SdNotify(true, daemon.SdNotifyReady)
	<-ctx.Done() // Server running
	daemon.SdNotify(true, daemon.SdNotifyStopping)

	// Graceful server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.Timeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.WithError(err).Error("shutting down http server")
	}

	err = dbh.Close()
	if err != nil {
		log.WithError(err).Error("closing database")
	}
}
