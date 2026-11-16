// Package applemusic holds the Apple Music authenticator.
package applemusic

import (
	"context"
	"fmt"
	"strings"

	"github.com/mathcale/setlist-to-playlist/internal/entities/applemusic"
	"github.com/mathcale/setlist-to-playlist/internal/entities/music"
	"github.com/mathcale/setlist-to-playlist/internal/infra/persistence"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/logger"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/prompter"
)

// StorefrontClient is the part of the Apple Music client the authenticator needs.
type StorefrontClient interface {
	SetAuth(auth applemusic.AuthData)
	Storefront(ctx context.Context) (string, error)
}

// Authenticator implements music.Authenticator for Apple Music. It loads
// saved tokens, asks for them on first use, and validates them with the
// storefront call. No web server is involved.
type Authenticator struct {
	persistence persistence.AppleMusicAuthPersistenceInterface
	client      StorefrontClient
	prompter    prompter.Prompter
	logger      logger.LoggerInterface
}

var _ music.Authenticator = (*Authenticator)(nil)

func NewAuthenticator(
	p persistence.AppleMusicAuthPersistenceInterface,
	c StorefrontClient,
	pr prompter.Prompter,
	l logger.LoggerInterface,
) *Authenticator {
	return &Authenticator{persistence: p, client: c, prompter: pr, logger: l}
}

func (a *Authenticator) Authenticate(ctx context.Context) error {
	auth, err := a.load()
	if err != nil {
		return err
	}

	a.client.SetAuth(*auth)

	storefront, err := a.client.Storefront(ctx)
	if err != nil {
		return err
	}

	a.logger.Info(fmt.Sprintf("Authenticated on Apple Music (storefront: %s)", storefront), nil)

	return nil
}

// load returns saved tokens, or asks for and saves them when missing.
func (a *Authenticator) load() (*applemusic.AuthData, error) {
	if saved, err := a.persistence.Read(); err == nil && saved.Validate() == nil {
		return saved, nil
	}

	bearer, err := a.prompter.Secret("What's your Apple Music web-player bearer token?")
	if err != nil {
		return nil, err
	}

	mut, err := a.prompter.Secret("What's your Apple Music media-user-token cookie value?")
	if err != nil {
		return nil, err
	}

	auth := applemusic.AuthData{
		BearerToken:    strings.TrimSpace(bearer),
		MediaUserToken: strings.TrimSpace(mut),
	}

	if err := auth.Validate(); err != nil {
		return nil, err
	}

	if err := a.persistence.Write(auth); err != nil {
		return nil, err
	}

	return &auth, nil
}
