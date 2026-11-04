package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTP               HTTPConfig
	PostgresDSN        string
	InvoiceSearchDSN   string
	RedisAddr          string
	SessionTTL         time.Duration
	ApprovalTTL        time.Duration
	LicenseTokenSecret string
	LicenseTokenTTL    time.Duration
	CookieSecure       bool
}

type HTTPConfig struct {
	Addr              string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
}

type lookupEnv func(string) (string, bool)

func Load() (Config, error) {
	return LoadFrom(os.LookupEnv)
}

func LoadFrom(lookup lookupEnv) (Config, error) {
	if lookup == nil {
		return Config{}, fmt.Errorf("load configuration: environment lookup is nil")
	}

	var cfg Config
	var problems []string
	cfg.PostgresDSN = required(lookup, "POSTGRES_DSN", &problems)
	cfg.InvoiceSearchDSN = optional(lookup, "INVOICE_SEARCH_DSN", defaultInvoiceSearchDSN(cfg.PostgresDSN))
	cfg.RedisAddr = required(lookup, "REDIS_ADDR", &problems)
	cfg.LicenseTokenSecret = optional(lookup, "LICENSE_TOKEN_SECRET", "wahrwelt-license-token-secret-key")
	if len([]byte(cfg.LicenseTokenSecret)) < 32 || strings.TrimSpace(cfg.LicenseTokenSecret) != cfg.LicenseTokenSecret {
		problems = append(problems, "LICENSE_TOKEN_SECRET must contain at least 32 bytes without surrounding whitespace")
	}
	if len(problems) > 0 {
		return Config{}, fmt.Errorf("configuration: %s", strings.Join(problems, ", "))
	}

	if err := validatePostgresDSN(cfg.PostgresDSN); err != nil {
		return Config{}, fmt.Errorf("POSTGRES_DSN %s", err)
	}
	if err := validatePostgresDSN(cfg.InvoiceSearchDSN); err != nil {
		return Config{}, fmt.Errorf("INVOICE_SEARCH_DSN %s", err)
	}
	if err := validateRedisAddr(cfg.RedisAddr); err != nil {
		return Config{}, fmt.Errorf("REDIS_ADDR %s", err)
	}

	cfg.HTTP.Addr = value(lookup, "HTTP_ADDR", ":8080")
	cfg.HTTP.ReadHeaderTimeout = duration(lookup, "HTTP_READ_HEADER_TIMEOUT", 2*time.Second, &problems)
	cfg.HTTP.ReadTimeout = duration(lookup, "HTTP_READ_TIMEOUT", 5*time.Second, &problems)
	cfg.HTTP.WriteTimeout = duration(lookup, "HTTP_WRITE_TIMEOUT", 10*time.Second, &problems)
	cfg.HTTP.IdleTimeout = duration(lookup, "HTTP_IDLE_TIMEOUT", 60*time.Second, &problems)
	cfg.HTTP.ShutdownTimeout = duration(lookup, "HTTP_SHUTDOWN_TIMEOUT", 10*time.Second, &problems)
	cfg.SessionTTL = duration(lookup, "SESSION_TTL", 12*time.Hour, &problems)
	cfg.ApprovalTTL = duration(lookup, "APPROVAL_TTL", 15*time.Minute, &problems)
	cfg.LicenseTokenTTL = duration(lookup, "LICENSE_TOKEN_TTL", 5*time.Minute, &problems)
	if cfg.LicenseTokenTTL < time.Minute || cfg.LicenseTokenTTL > 15*time.Minute {
		problems = append(problems, "LICENSE_TOKEN_TTL must be between 1m and 15m")
	}
	cfg.CookieSecure = boolean(lookup, "COOKIE_SECURE", false, &problems)
	if len(problems) > 0 {
		return Config{}, fmt.Errorf("configuration: %s", strings.Join(problems, ", "))
	}

	if strings.TrimSpace(cfg.HTTP.Addr) == "" {
		return Config{}, fmt.Errorf("HTTP_ADDR must not be empty")
	}
	return cfg, nil
}

func optional(lookup lookupEnv, key, fallback string) string {
	if value, ok := lookup(key); ok && strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func defaultInvoiceSearchDSN(dsn string) string {
	parsed, err := url.Parse(dsn)
	if err != nil || parsed.Hostname() == "" {
		return "postgres://wahrwelt_invoice_search:wahrwelt@postgres:5432/wahrwelt?sslmode=disable"
	}
	parsed.User = url.UserPassword("wahrwelt_invoice_search", "wahrwelt")
	return parsed.String()
}

func required(lookup lookupEnv, key string, problems *[]string) string {
	value, ok := lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		*problems = append(*problems, key+" is required")
		return ""
	}
	return value
}

func value(lookup lookupEnv, key, fallback string) string {
	if value, ok := lookup(key); ok {
		return value
	}
	return fallback
}

func duration(lookup lookupEnv, key string, fallback time.Duration, problems *[]string) time.Duration {
	text, ok := lookup(key)
	if !ok {
		return fallback
	}
	parsed, err := time.ParseDuration(text)
	if err != nil || parsed <= 0 {
		*problems = append(*problems, key+" must be a positive duration")
		return 0
	}
	return parsed
}

func boolean(lookup lookupEnv, key string, fallback bool, problems *[]string) bool {
	text, ok := lookup(key)
	if !ok {
		return fallback
	}
	parsed, err := strconv.ParseBool(text)
	if err != nil {
		*problems = append(*problems, key+" must be true or false")
		return false
	}
	return parsed
}

func validatePostgresDSN(dsn string) error {
	parsed, err := url.Parse(dsn)
	if err != nil || (parsed.Scheme != "postgres" && parsed.Scheme != "postgresql") || parsed.Hostname() == "" || strings.Trim(parsed.Path, "/") == "" {
		return fmt.Errorf("must be a valid PostgreSQL URL with a host and database")
	}
	if parsed.User == nil || strings.TrimSpace(parsed.User.Username()) == "" {
		return fmt.Errorf("must include a username")
	}
	return nil
}

func validateRedisAddr(address string) error {
	host, portText, err := net.SplitHostPort(address)
	if err != nil || strings.TrimSpace(host) == "" {
		return fmt.Errorf("must be a host and port")
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("must use a port from 1 through 65535")
	}
	return nil
}
