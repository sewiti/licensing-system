package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	mathrand "math/rand"

	"github.com/apex/log"
	"github.com/coreos/go-systemd/daemon"
	"github.com/gorilla/handlers"
	"github.com/sewiti/licensing-system/internal/core"
	"github.com/sewiti/licensing-system/internal/db"
	"github.com/sewiti/licensing-system/internal/server"
	"github.com/vrischmann/envconfig"
)

func runServer() error {
	mathrand.Seed(time.Now().UnixNano())

	// Config
	var cfg config
	err := envconfig.Init(&cfg)
	if err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	wg := sync.WaitGroup{}

	// Database
	migrated, err := db.MigrateUp(cfg.DbDSN)
	if migrated {
		log.Info("migrated database")
	}
	if err != nil {
		return fmt.Errorf("migrating db: %w", err)
	}
	db, err := db.Open(cfg.DbDSN)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}

	// Core
	conf := core.LicensingConf{
		Limiter:          core.LimiterConf(cfg.Licensing.Limiter),
		Refresh:          core.RefreshConf(cfg.Licensing.Refresh),
		MaxTimeDrift:     cfg.Licensing.MaxTimeDrift,
		MinPasswdEntropy: cfg.MinPasswdEntropy,
	}
	c, err := core.NewCore(db, cfg.Licensing.ServerKey, time.Now(), conf)
	if err != nil {
		return fmt.Errorf("create runtime: %w", err)
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

	// Server
	r := server.NewRouter(c,
		cfg.HTTP.CORS.ResourceApiEnabled,
		cfg.HTTP.CORS.LicensingApiEnabled,
		cfg.HTTP.CORS.AllowedOrigins)
	r.Use(handlers.CompressHandler)
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
			log.WithError(err).Error("listening and serving server")
		}
	}()

	// Internal Server (for CLI)
	err = os.Remove(cfg.InternalSocket)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("internal server: %w", err)
		}
	}
	l, err := net.ListenUnix("unix", &net.UnixAddr{Name: cfg.InternalSocket, Net: "unix"})
	if err != nil {
		return fmt.Errorf("internal server: %w", err)
	}
	defer os.Remove(cfg.InternalSocket)
	err = os.Chmod(cfg.InternalSocket, 0700)
	if err != nil {
		return fmt.Errorf("internal server: %w", err)
	}
	srvi := http.Server{
		Handler:      server.NewRouterInternal(c),
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}
	go func() {
		defer cancel()
		err := srvi.Serve(l)
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			log.WithError(err).Error("serving internal server")
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
		return fmt.Errorf("close db: %w", err)
	}
	return nil
}
