package config

import (
	"os"
	"path"

	"github.com/charmbracelet/huh"
	"github.com/spf13/viper"

	"github.com/mathcale/setlist-to-playlist/internal/infra/persistence/drivers"
)

type General struct {
	LogLevel      string `mapstructure:"log_level"`
	WebServerPort int64  `mapstructure:"webserver_port"`
}

type SetlistFM struct {
	APIKey  string `mapstructure:"api_key"`
	BaseURL string `mapstructure:"base_url"`
	Timeout int    `mapstructure:"timeout_ms"`
}

type Spotify struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
}

type Config struct {
	General   `mapstructure:"general"`
	SetlistFM `mapstructure:"setlistfm"`
	Spotify   `mapstructure:"spotify"`
}

type ConfigPaths struct {
	AppConfigDir    string
	AppConfigFile   string
	SpotifyAuthFile string
}

func Init() (*ConfigPaths, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	fsDriver := drivers.NewFileSystemDriver()

	appConfigDirPath := path.Join(userConfigDir, "setlist-to-playlist")
	appConfigFilePath := path.Join(appConfigDirPath, "config.toml")
	spotifyAuthFilePath := path.Join(appConfigDirPath, "spotify_auth.json")

	if err := fsDriver.CreateDir(appConfigDirPath, 0750); err != nil {
		return nil, err
	}

	if exists := fsDriver.Exists(spotifyAuthFilePath); !exists {
		if err := fsDriver.Write(spotifyAuthFilePath, []byte("{}"), 0660); err != nil {
			return nil, err
		}
	}

	return &ConfigPaths{
		AppConfigDir:    appConfigDirPath,
		AppConfigFile:   appConfigFilePath,
		SpotifyAuthFile: spotifyAuthFilePath,
	}, nil
}

func Load(configPaths ConfigPaths) (*Config, error) {
	var c *Config

	viper.WatchConfig()
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(configPaths.AppConfigDir)
	viper.AddConfigPath(".")

	viper.SetDefault("general.log_level", "info")
	viper.SetDefault("general.webserver_port", 8080)
	viper.SetDefault("setlistfm.base_url", "https://api.setlist.fm/rest")
	viper.SetDefault("setlistfm.timeout_ms", 3000)
	viper.SetDefault("spotify.redirect_url", "http://localhost:8080/callback")

	if ok := viper.IsSet("setlistfm.api_key"); !ok {
		var apiKey string
		huh.NewInput().Title("What's your Setlist.fm API key?").Prompt(">").Value(&apiKey).Run()

		viper.Set("setlistfm.api_key", apiKey)
		viper.WriteConfig()
	}

	if ok := viper.IsSet("spotify.client_id"); !ok {
		var clientID string
		huh.NewInput().Title("What's your Spotify client ID?").Prompt(">").Value(&clientID).Run()

		viper.Set("spotify.client_id", clientID)
		viper.WriteConfig()
	}

	if ok := viper.IsSet("spotify.client_secret"); !ok {
		var secret string
		huh.NewInput().Title("What's your Spotify client secret?").Prompt(">").Value(&secret).Run()

		viper.Set("spotify.client_secret", secret)
		viper.WriteConfig()
	}

	if err := viper.SafeWriteConfigAs(configPaths.AppConfigFile); err != nil {
		if _, ok := err.(viper.ConfigFileAlreadyExistsError); !ok {
			return nil, err
		}
	}

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := viper.Unmarshal(&c); err != nil {
		return nil, err
	}

	return c, nil
}
