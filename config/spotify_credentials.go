package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"

	"github.com/mathcale/setlist-to-playlist/internal/pkg/prompter"
)

const (
	spotifyClientIDKey     = "spotify.client_id"
	spotifyClientSecretKey = "spotify.client_secret"
)

var credentialLabels = map[string]string{
	spotifyClientIDKey:     "Spotify client ID",
	spotifyClientSecretKey: "Spotify client secret",
}

// SpotifyCredentialStore asks for the Spotify client ID and secret on first
// use and persists them to the TOML config. Load must have run before.
type SpotifyCredentialStore struct {
	prompter prompter.Prompter
}

func NewSpotifyCredentialStore(p prompter.Prompter) *SpotifyCredentialStore {
	return &SpotifyCredentialStore{prompter: p}
}

// Ensure prompts only for credentials that are not yet saved.
func (s *SpotifyCredentialStore) Ensure() error {
	if viper.GetString(spotifyClientIDKey) == "" {
		id, err := s.prompter.Text("What's your Spotify client ID?")
		if err != nil {
			return err
		}

		if err := save(spotifyClientIDKey, id); err != nil {
			return err
		}
	}

	if viper.GetString(spotifyClientSecretKey) == "" {
		secret, err := s.prompter.Secret("What's your Spotify client secret?")
		if err != nil {
			return err
		}

		if err := save(spotifyClientSecretKey, secret); err != nil {
			return err
		}
	}

	return nil
}

// Credentials returns the saved credentials (empty until Ensure succeeded).
func (s *SpotifyCredentialStore) Credentials() (string, string) {
	return viper.GetString(spotifyClientIDKey), viper.GetString(spotifyClientSecretKey)
}

func save(key, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%s cannot be empty", credentialLabels[key])
	}

	viper.Set(key, value)
	return viper.WriteConfig()
}
