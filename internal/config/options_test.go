package config

import (
	"flag"
	"os"
	"testing"
)

func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	Options = struct {
		AppPort      string `default:"localhost:8080"`
		CertFile     string
		KeyFile      string
		DatabaseDSN  string
		JWTSecretKey string
	}{}
}

func TestParseFlags_Defaults(t *testing.T) {
	resetFlags()
	os.Args = []string{"cmd"}

	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("DATABASE_DSN")
	os.Unsetenv("JWT_SECRET")

	ParseFlags()

	if Options.AppPort != "localhost:8080" {
		t.Errorf("expected default AppPort 'localhost:8080', got %q", Options.AppPort)
	}
	if Options.DatabaseDSN != "" {
		t.Errorf("expected empty DatabaseDSN, got %q", Options.DatabaseDSN)
	}
}

func TestParseFlags_CLIArgs(t *testing.T) {
	resetFlags()
	os.Args = []string{"cmd", "-a", "0.0.0.0:9090", "-d", "postgres://localhost/testdb", "-j", "mysecret"}

	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("DATABASE_DSN")
	os.Unsetenv("JWT_SECRET")

	ParseFlags()

	if Options.AppPort != "0.0.0.0:9090" {
		t.Errorf("expected AppPort '0.0.0.0:9090', got %q", Options.AppPort)
	}
	if Options.DatabaseDSN != "postgres://localhost/testdb" {
		t.Errorf("expected DatabaseDSN 'postgres://localhost/testdb', got %q", Options.DatabaseDSN)
	}
	if Options.JWTSecretKey != "mysecret" {
		t.Errorf("expected JWTSecretKey 'mysecret', got %q", Options.JWTSecretKey)
	}
}

func TestParseFlags_EnvOverridesCLI(t *testing.T) {
	resetFlags()
	os.Args = []string{"cmd", "-a", "localhost:8080", "-d", "cli-dsn", "-j", "cli-secret"}

	t.Setenv("SERVER_ADDRESS", "env-addr:3000")
	t.Setenv("DATABASE_DSN", "env-dsn")
	t.Setenv("JWT_SECRET", "env-secret")

	ParseFlags()

	if Options.AppPort != "env-addr:3000" {
		t.Errorf("expected env AppPort 'env-addr:3000', got %q", Options.AppPort)
	}
	if Options.DatabaseDSN != "env-dsn" {
		t.Errorf("expected env DatabaseDSN 'env-dsn', got %q", Options.DatabaseDSN)
	}
	if Options.JWTSecretKey != "env-secret" {
		t.Errorf("expected env JWTSecretKey 'env-secret', got %q", Options.JWTSecretKey)
	}
}

func TestParseFlags_PartialEnvOverride(t *testing.T) {
	resetFlags()
	os.Args = []string{"cmd", "-a", "cli-addr:8080", "-d", "cli-dsn"}

	t.Setenv("SERVER_ADDRESS", "env-addr:9090")
	os.Unsetenv("DATABASE_DSN")
	os.Unsetenv("JWT_SECRET")

	ParseFlags()

	if Options.AppPort != "env-addr:9090" {
		t.Errorf("expected env AppPort 'env-addr:9090', got %q", Options.AppPort)
	}
	if Options.DatabaseDSN != "cli-dsn" {
		t.Errorf("expected CLI DatabaseDSN 'cli-dsn', got %q", Options.DatabaseDSN)
	}
}
