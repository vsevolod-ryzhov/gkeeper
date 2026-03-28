package config

import (
	"flag"
	"os"
)

var Options struct {
	AppPort      string `default:"localhost:8080"`
	CertFile     string
	KeyFile      string
	DatabaseDSN  string
	JWTSecretKey string
}

func ParseFlags() {
	flag.StringVar(&Options.AppPort, "a", "localhost:8080", "The address to bind the app to")
	flag.StringVar(&Options.CertFile, "c", "", "The TLS certificate file")
	flag.StringVar(&Options.KeyFile, "k", "", "The TLS key file")
	flag.StringVar(&Options.DatabaseDSN, "d", "", "Database connection string")
	flag.StringVar(&Options.JWTSecretKey, "j", "", "JWT secret key")
	flag.Parse()

	if envRunAddr, exists := os.LookupEnv("SERVER_ADDRESS"); exists {
		Options.AppPort = envRunAddr
	}

	if envDatabaseDSN, exists := os.LookupEnv("DATABASE_DSN"); exists {
		Options.DatabaseDSN = envDatabaseDSN
	}

	if envJWTSecret, exists := os.LookupEnv("JWT_SECRET"); exists {
		Options.JWTSecretKey = envJWTSecret
	}

	if certFile, exists := os.LookupEnv("SERVER_CERT"); exists {
		Options.CertFile = certFile
	}

	if keyFile, exists := os.LookupEnv("SERVER_KEY"); exists {
		Options.KeyFile = keyFile
	}
}
