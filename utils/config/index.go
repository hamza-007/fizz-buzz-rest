package config

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"

	viper "github.com/spf13/viper"
)

func Environment() string { return viper.GetString("ENVIRONMENT") }

func IsProduction() bool { return Environment() == "production" }

func IsStaging() bool { return Environment() == "staging" }

func IsDevelopment() bool { return Environment() == "development" || Environment() == "" }

func IsOnline() bool { return IsProduction() || IsStaging() }

func APP() app { return current.appConfig }

func FizzBuzz() fizzbuzz { return current.fbConfig }

type container struct {
	appConfig app
	fbConfig  fizzbuzz
}

var current = &container{}

func Load() error {
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	for _, file := range []string{".env", ".env.local"} {
		if err := mergeEnvFile(file); err != nil {
			return err
		}
	}

	loaded := &container{}
	if err := loaded.appConfig.load(); err != nil {
		return err
	}
	if err := loaded.fbConfig.load(); err != nil {
		return err
	}

	current = loaded
	return nil
}

func mergeEnvFile(file string) error {
	viper.SetConfigType("env")
	viper.SetConfigFile(file)

	err := viper.MergeInConfig()
	switch {
	case err == nil,
		errors.As(err, &viper.ConfigFileNotFoundError{}),
		errors.Is(err, fs.ErrNotExist):
		return nil
	default:
		return fmt.Errorf("loading %s: %w", file, err)
	}
}
