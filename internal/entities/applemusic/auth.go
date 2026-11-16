// Package applemusic holds the Apple Music specific domain data.
package applemusic

import (
	"errors"
	"strings"
)

// AuthData is what the CLI needs to talk to Apple Music: the web-player
// bearer (developer) token and the media-user-token cookie value.
type AuthData struct {
	BearerToken    string `json:"bearer_token"`
	MediaUserToken string `json:"media_user_token"`
}

// Validate rejects missing tokens.
func (d AuthData) Validate() error {
	if strings.TrimSpace(d.BearerToken) == "" {
		return errors.New("bearer token cannot be empty")
	}

	if strings.TrimSpace(d.MediaUserToken) == "" {
		return errors.New("media-user-token cannot be empty")
	}

	return nil
}
