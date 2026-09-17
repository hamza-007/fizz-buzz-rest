package config

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	viper "github.com/spf13/viper"
)

type app struct {
	Port              int
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	RequestTimeout    time.Duration
	LogLevel          slog.Level
	CORS              []string
	TrustProxy        bool
	RateLimit         int
	RateWindow        time.Duration
}

func (app) namespace() string { return "APP" }

func (obj app) key(name string) string { return fmt.Sprintf("%s_%s", obj.namespace(), name) }

func (obj app) Addr() string { return fmt.Sprintf(":%d", obj.Port) }

func (obj *app) load() error {
	viper.SetDefault(obj.key("PORT"), 8080)
	viper.SetDefault(obj.key("READ_HEADER_TIMEOUT"), 5*time.Second)
	viper.SetDefault(obj.key("READ_TIMEOUT"), 10*time.Second)
	viper.SetDefault(obj.key("WRITE_TIMEOUT"), 30*time.Second)
	viper.SetDefault(obj.key("IDLE_TIMEOUT"), 60*time.Second)
	viper.SetDefault(obj.key("SHUTDOWN_TIMEOUT"), 10*time.Second)
	viper.SetDefault(obj.key("REQUEST_TIMEOUT"), 20*time.Second)
	viper.SetDefault(obj.key("LOG_LEVEL"), "info")
	viper.SetDefault(obj.key("TRUST_PROXY"), false)
	viper.SetDefault(obj.key("RATE_LIMIT"), 100)
	viper.SetDefault(obj.key("RATE_WINDOW"), time.Minute)

	obj.Port = viper.GetInt(obj.key("PORT"))
	obj.ReadHeaderTimeout = viper.GetDuration(obj.key("READ_HEADER_TIMEOUT"))
	obj.ReadTimeout = viper.GetDuration(obj.key("READ_TIMEOUT"))
	obj.WriteTimeout = viper.GetDuration(obj.key("WRITE_TIMEOUT"))
	obj.IdleTimeout = viper.GetDuration(obj.key("IDLE_TIMEOUT"))
	obj.ShutdownTimeout = viper.GetDuration(obj.key("SHUTDOWN_TIMEOUT"))
	obj.RequestTimeout = viper.GetDuration(obj.key("REQUEST_TIMEOUT"))
	obj.CORS = splitList(viper.GetString(obj.key("CORS")))
	obj.TrustProxy = viper.GetBool(obj.key("TRUST_PROXY"))
	obj.RateLimit = viper.GetInt(obj.key("RATE_LIMIT"))
	obj.RateWindow = viper.GetDuration(obj.key("RATE_WINDOW"))

	if viper.GetBool("verbose") {
		obj.LogLevel = slog.LevelDebug
	} else if err := obj.LogLevel.UnmarshalText([]byte(viper.GetString(obj.key("LOG_LEVEL")))); err != nil {
		return fmt.Errorf("%s: %q is not a valid level (debug, info, warn, error)", obj.key("LOG_LEVEL"), viper.GetString(obj.key("LOG_LEVEL")))
	}

	return obj.validate()
}

func splitList(raw string) []string {
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			values = append(values, part)
		}
	}
	return values
}

func (obj app) validate() error {
	if obj.Port < 1 || obj.Port > 65535 {
		return fmt.Errorf("%s: %d is not a valid port", obj.key("PORT"), obj.Port)
	}

	durations := map[string]time.Duration{
		obj.key("READ_HEADER_TIMEOUT"): obj.ReadHeaderTimeout,
		obj.key("READ_TIMEOUT"):        obj.ReadTimeout,
		obj.key("WRITE_TIMEOUT"):       obj.WriteTimeout,
		obj.key("IDLE_TIMEOUT"):        obj.IdleTimeout,
		obj.key("SHUTDOWN_TIMEOUT"):    obj.ShutdownTimeout,
		obj.key("REQUEST_TIMEOUT"):     obj.RequestTimeout,
		obj.key("RATE_WINDOW"):         obj.RateWindow,
	}
	for key, value := range durations {
		if value <= 0 {
			return fmt.Errorf("%s: must be a positive duration (e.g. 10s, 1m)", key)
		}
	}

	if obj.RateLimit < 0 {
		return fmt.Errorf("%s: must not be negative (0 disables throttling)", obj.key("RATE_LIMIT"))
	}

	if obj.RequestTimeout >= obj.WriteTimeout {
		return fmt.Errorf("%s: must be shorter than %s", obj.key("REQUEST_TIMEOUT"), obj.key("WRITE_TIMEOUT"))
	}

	return nil
}
