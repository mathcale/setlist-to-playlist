package config

import "github.com/spf13/viper"

type Config struct {
	SetlistFMAPIKey     string `mapstructure:"SETLISTFM_API_KEY"`
	SetlistFMAPIBaseURL string `mapstructure:"SETLISTFM_API_BASE_URL"`
	SetlistFMAPITimeout int    `mapstructure:"SETLISTFM_API_TIMEOUT"`
}

func Load(path string) (*Config, error) {
	var c *Config

	viper.SetConfigName("app_config")
	viper.SetConfigType("env")
	viper.AddConfigPath(path)
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := viper.Unmarshal(&c); err != nil {
		return nil, err
	}

	return c, nil
}
