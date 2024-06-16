package spotify

import (
	"errors"
	"time"

	"golang.org/x/oauth2"
)

type SpotifyUserAuthData struct {
	AccessToken  string `json:"access_token"`
	Expiry       string `json:"expiry"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

func NewSpotifyUserAuthData(
	accessToken string,
	expiry string,
	refreshToken string,
	tokenType string,
) SpotifyUserAuthData {
	return SpotifyUserAuthData{
		AccessToken:  accessToken,
		Expiry:       expiry,
		RefreshToken: refreshToken,
		TokenType:    tokenType,
	}
}

func (data *SpotifyUserAuthData) Validate() error {
	if data == nil {
		return errors.New("no authentication data found")
	}

	if data.AccessToken == "" {
		return errors.New("no token found in authentication data")
	}

	if data.RefreshToken == "" {
		return errors.New("no refresh token found in authentication data")
	}

	if data.Expiry == "" {
		return errors.New("no expiry date found in authentication data")
	}

	exp, _ := time.Parse(time.RFC3339, data.Expiry)
	if time.Now().After(exp) {
		return errors.New("token has expired")
	}

	return nil
}

func (d SpotifyUserAuthData) ToOauth2Token() (*oauth2.Token, error) {
	exp, err := time.Parse(time.RFC3339, d.Expiry)
	if err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken:  d.AccessToken,
		Expiry:       exp,
		RefreshToken: d.RefreshToken,
		TokenType:    d.TokenType,
	}, nil
}
