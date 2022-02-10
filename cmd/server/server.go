package main

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/apex/log"
	"github.com/coreos/go-systemd/daemon"
	"github.com/gorilla/handlers"
	"github.com/sewiti/licensing-system/internal/core"
	"github.com/sewiti/licensing-system/internal/db"
	"github.com/sewiti/licensing-system/internal/server"
	"github.com/vrischmann/envconfig"
)

func runServer() {
	// Config
	var cfg config
	err := envconfig.Init(&cfg)
	if err != nil {
		log.WithError(err).Fatal("initializing environment configuration")
		return
	}

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	wg := sync.WaitGroup{}

	// Database
	migrated, err := db.MigrateUp(cfg.DbDataSource)
	if migrated {
		log.Info("migrated database")
	}
	if err != nil {
		log.WithError(err).Fatal("migrating database")
		return
	}
	db, err := db.Open(cfg.DbDataSource)
	if err != nil {
		log.WithError(err).Fatal("opening database")
		return
	}

	// Core
	conf := core.LicensingConf{
		Limiter:      core.LimiterConf(cfg.Licensing.Limiter),
		Refresh:      core.RefreshConf(cfg.Licensing.Refresh),
		MaxTimeDrift: cfg.Licensing.MaxTimeDrift,
	}
	c, err := core.NewCore(db, cfg.Licensing.ServerKey, time.Now(), conf)
	if err != nil {
		log.WithError(err).Fatal("creating runtime")
		return
	}
	if cfg.Licensing.CleanupInterval > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			core.RunCleanupRoutine(ctx, db, cfg.Licensing.CleanupInterval, func(msg string, err error) {
				if err != nil {
					log.WithError(err).Error(msg)
				} else {
					log.Info(msg)
				}
			})
		}()
	}

	// Router
	r := server.NewRouter(c,
		cfg.HTTP.CORS.Enabled,
		cfg.HTTP.CORS.AllowedOrigins)
	r.Use(handlers.CompressHandler)

	// Server
	srv := http.Server{
		Addr:         cfg.HTTP.Listen,
		Handler:      r,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}
	go func() {
		defer cancel()
		var err error
		if cfg.HTTP.TLS.CertFile == "" && cfg.HTTP.TLS.KeyFile == "" {
			err = srv.ListenAndServe()
		} else {
			err = srv.ListenAndServeTLS(cfg.HTTP.TLS.CertFile, cfg.HTTP.TLS.KeyFile)
		}
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
	ctx, cancel = context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	err = srv.Shutdown(ctx)
	if err != nil {
		log.WithError(err).Error("shutting down http server")
	}

	wg.Wait()

	err = db.Close()
	if err != nil {
		log.WithError(err).Error("closing database")
	}
}
