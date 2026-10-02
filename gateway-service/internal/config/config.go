package config

import (
	"os"
)

type Config struct {
	Port       string
	CarsURL    string
	RentalURL  string
	PaymentURL string
}

func Load() *Config {
	return &Config{
		Port:       getEnv("PORT", "8050"),
		CarsURL:    getEnv("CARS_URL", "http://localhost:8070"),
		RentalURL:  getEnv("RENTAL_URL", "http://localhost:8060"),
		PaymentURL: getEnv("PAYMENT_URL", "http://localhost:8050"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
