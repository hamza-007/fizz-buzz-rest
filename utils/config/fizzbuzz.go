package config

import (
	"fmt"

	viper "github.com/spf13/viper"
)

type fizzbuzz struct {
	MaxLimit     int
	MaxStatsKeys int
}

func (fizzbuzz) namespace() string          { return "FIZZBUZZ" }
func (obj fizzbuzz) key(name string) string { return fmt.Sprintf("%s_%s", obj.namespace(), name) }

func (obj *fizzbuzz) load() error {
	viper.SetDefault(obj.key("MAX_LIMIT"), 100_000)
	viper.SetDefault(obj.key("MAX_STATS_KEYS"), 10_000)

	obj.MaxLimit = viper.GetInt(obj.key("MAX_LIMIT"))
	obj.MaxStatsKeys = viper.GetInt(obj.key("MAX_STATS_KEYS"))

	return obj.validate()
}

func (obj fizzbuzz) validate() error {

	if obj.MaxLimit < 1 {
		return fmt.Errorf("%s: must be a positive integer", obj.key("MAX_LIMIT"))
	}
	if obj.MaxStatsKeys < 0 {
		return fmt.Errorf("%s: must not be negative (0 means unbounded)", obj.key("MAX_STATS_KEYS"))
	}
	return nil
}
