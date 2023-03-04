package config_parser

import (
	"bytes"
	"errors"
	"github.com/ihatiko/go-chef/pkg/config"
	"github.com/ihatiko/log"
	"github.com/spf13/viper"
)

const (
	yml = "yml"
)

func LoadConfig(file []byte) (*config.Config, error) {
	cfg := config.New(viper.New())
	cfg.AddConfigPath(".")
	cfg.AutomaticEnv()
	cfg.SetConfigType(yml)
	if err := cfg.ReadConfig(bytes.NewReader(file)); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, errors.New("config file not found")
		}

		return nil, err
	}

	return cfg, nil
}
func ParseConfig[T any](v *config.Config) (*T, error) {
	var c T

	err := v.Unmarshal(&c)
	if err != nil {
		log.Error("unable to decode into struct, %v", err)
		return nil, err
	}

	return &c, nil
}
