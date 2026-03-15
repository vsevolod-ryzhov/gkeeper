package config

import (
	"flag"
	"os"
)

var Options struct {
	AppPort     string `default:"localhost:8080"`
	DatabaseDSN string
}

func ParseFlags() {
	flag.StringVar(&Options.AppPort, "a", "localhost:8080", "The address to bind the app to")
	flag.StringVar(&Options.AppPort, "d", "", "Database connection string")
	flag.Parse()

	if envRunAddr, exists := os.LookupEnv("SERVER_ADDRESS"); exists {
		Options.AppPort = envRunAddr
	}
}
