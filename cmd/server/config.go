package main

import "time"

type config struct {
	HTTP struct {
		Listen  string        `envconfig:"default=:1555"`
		Timeout time.Duration `envconfig:"default=30s"`

		CORS struct {
			AllowedHeaders []string `envconfig:"default=Authorization;Content-Type"`
			AllowedMethods []string `envconfig:"default=GET;POST;PUT;DELETE"`
			AllowedOrigins []string `envconfig:"optional"`
		}
	}
	DB struct {
		DataSource string
	}
}
