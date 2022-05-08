package main

import "time"

// config defines licensing server config.
type config struct {
	DbDSN string

	HTTP struct {
		Listen          string        `envconfig:"optional"`
		ReadTimeout     time.Duration `envconfig:"default=30s"`
		WriteTimeout    time.Duration `envconfig:"default=30s"`
		ShutdownTimeout time.Duration `envconfig:"default=30s"`
		Gzip            bool          `envconfig:"default=false"`

		CORS struct {
			ResourceApiEnabled  bool     `envconfig:"default=false"`
			LicensingApiEnabled bool     `envconfig:"default=false"`
			AllowedOrigins      []string `envconfig:"optional"`
		}

		TLS struct {
			CertFile string `envconfig:"optional"`
			KeyFile  string `envconfig:"optional"`
		}
	}

	Licensing struct {
		ServerKey       []byte
		MaxTimeDrift    time.Duration `envconfig:"default=6h"`
		CleanupInterval time.Duration `envconfig:"default=20m"`

		Refresh struct {
			Min    time.Duration `envconfig:"default=5m"`
			Max    time.Duration `envconfig:"default=2h"`
			Jitter float64       `envconfig:"default=0.1"`
		}

		Limiter struct {
			SessionEvery     time.Duration `envconfig:"default=10m"`
			SessionEveryInit time.Duration `envconfig:"default=1m"` // not used due to a bug
			BurstTotal       time.Duration `envconfig:"default=8h"`

			CacheExpiration      time.Duration `envconfig:"default=24h"`
			CacheCleanupInterval time.Duration `envconfig:"default=1h"`
		}
	}

	InternalSocket   string  `envconfig:"default=/run/licensing-server.sock"`
	MinPasswdEntropy float64 `envconfig:"default=30"`

	DisableGUI bool `envconfig:"default=false"`
}
