package main

import "time"

type config struct {
	DbDataSource string

	HTTP struct {
		Listen          string        `envconfig:"optional"`
		ReadTimeout     time.Duration `envconfig:"default=30s"`
		WriteTimeout    time.Duration `envconfig:"default=30s"`
		ShutdownTimeout time.Duration `envconfig:"default=30s"`

		CORS struct {
			Enabled        bool     `envconfig:"default=false"`
			AllowedOrigins []string `envconfig:"optional"`
		}

		TLS struct {
			CertFile string `envconfig:"optional"`
			KeyFile  string `envconfig:"optional"`
		}
	}

	Licensing struct {
		ServerKey       []byte
		MaxTimeDrift    time.Duration `envconfig:"default=24h"`
		CleanupInterval time.Duration `envconfig:"default=30m"`

		Refresh struct {
			Min    time.Duration `envconfig:"default=5m"`
			Max    time.Duration `envconfig:"default=4h"`
			Jitter float64       `envconfig:"default=0.1"`
		}

		Limiter struct {
			SessionEvery     time.Duration `envconfig:"default=10m"`
			SessionEveryInit time.Duration `envconfig:"default=1m"`
			BurstTotal       time.Duration `envconfig:"default=8h"`

			CacheExpiration      time.Duration `envconfig:"default=24h"`
			CacheCleanupInterval time.Duration `envconfig:"default=1h"`
		}
	}
}
