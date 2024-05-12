package config

import "github.com/spf13/viper"

type Config struct {
	SetlistFMAPIKey string `mapstructure:"SETLISTFM_API_KEY"`
}

func Load(path string) (*Config, error) {
	var c *Config

	viper.SetConfigName("app_config")
	viper.SetConfigType("env")
	viper.AddConfigPath(path)
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	if err := viper.Unmarshal(&c); err != nil {
		panic(err)
	}

	return c, nil
}
