package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	*viper.Viper
}

func New(v *viper.Viper) *Config {
	return &Config{v}
}
