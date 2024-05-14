package config

import "github.com/spf13/viper"

type Config struct {
	LogLevel            string `mapstructure:"LOG_LEVEL"`
	WebServerPort       int64  `mapstructure:"WEBSERVER_PORT"`
	SetlistFMAPIKey     string `mapstructure:"SETLISTFM_API_KEY"`
	SetlistFMAPIBaseURL string `mapstructure:"SETLISTFM_API_BASE_URL"`
	SetlistFMAPITimeout int    `mapstructure:"SETLISTFM_API_TIMEOUT"`
	SpotifyClientID     string `mapstructure:"SPOTIFY_ID"`
	SpotifyClientSecret string `mapstructure:"SPOTIFY_SECRET"`
	SpotifyRedirectURL  string `mapstructure:"SPOTIFY_REDIRECT_URL"`
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
