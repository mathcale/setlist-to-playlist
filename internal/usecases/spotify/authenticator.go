package spotify

import (
	"context"
	"sync"

	"github.com/mathcale/setlist-to-playlist/internal/entities/music"
	oauth2util "github.com/mathcale/setlist-to-playlist/internal/pkg/oauth2"
)

// CredentialsEnsurer makes sure the Spotify app credentials exist, asking the
// user for them on first use.
type CredentialsEnsurer interface {
	Ensure() error
}

// WebServerStarter starts the local server that receives the OAuth callback.
type WebServerStarter interface {
	Start()
}

// SpotifyAuthenticator implements music.Authenticator for Spotify: it ensures
// the app credentials exist, starts the callback web server and runs the
// existing PKCE flow (refreshing saved auth data when still valid).
type SpotifyAuthenticator struct {
	credentials CredentialsEnsurer
	webServer   WebServerStarter
	useCase     SpotifyUserAuthenticationUseCaseInterface
	pkceCodes   oauth2util.GenerateOutput
	state       string

	startServer sync.Once
}

var _ music.Authenticator = (*SpotifyAuthenticator)(nil)

func NewSpotifyAuthenticator(
	credentials CredentialsEnsurer,
	webServer WebServerStarter,
	useCase SpotifyUserAuthenticationUseCaseInterface,
	pkceCodes oauth2util.GenerateOutput,
	state string,
) *SpotifyAuthenticator {
	return &SpotifyAuthenticator{
		credentials: credentials,
		webServer:   webServer,
		useCase:     useCase,
		pkceCodes:   pkceCodes,
		state:       state,
	}
}

func (a *SpotifyAuthenticator) Authenticate(ctx context.Context) error {
	if err := a.credentials.Ensure(); err != nil {
		return err
	}

	a.startServer.Do(a.webServer.Start)

	return a.useCase.Execute(ctx, a.pkceCodes, a.state)
}
