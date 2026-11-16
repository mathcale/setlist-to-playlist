package gateways

import (
	"context"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mathcale/setlist-to-playlist/config"
	"github.com/mathcale/setlist-to-playlist/internal/entities/music"
	oauth2util "github.com/mathcale/setlist-to-playlist/internal/pkg/oauth2"
	"github.com/mathcale/setlist-to-playlist/internal/tests/mocks"
	applemusic_ucs "github.com/mathcale/setlist-to-playlist/internal/usecases/applemusic"
	spotify_ucs "github.com/mathcale/setlist-to-playlist/internal/usecases/spotify"
)

// Each provider must ask only for its own credentials. This wires the real
// Spotify credential store + authenticator and the real Apple authenticator
// behind the real gateway/registry, sharing one PrompterMock. Only the
// OAuth/web-server edge (use case + server) and the Apple HTTP client are
// faked, since the PKCE flow needs a browser.
type credentialsFixture struct {
	prompter *mocks.PrompterMock
	gateway  RootCmdGatewayInterface
}

func newCredentialsFixture(t *testing.T) *credentialsFixture {
	t.Helper()

	viper.Reset()
	t.Cleanup(viper.Reset)

	cfgFile := t.TempDir() + "/config.toml"
	assert.NoError(t, os.WriteFile(cfgFile, []byte("[setlistfm]\napi_key='k'\n"), 0600))
	viper.SetConfigFile(cfgFile)
	assert.NoError(t, viper.ReadInConfig())

	p := new(mocks.PrompterMock)

	webServer := new(mocks.WebServerMock)
	webServer.On("Start").Return()

	spotifyFlow := new(mocks.SpotifyUserAuthenticationUseCaseMock)
	spotifyFlow.On("Execute", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	spotifyAuth := spotify_ucs.NewSpotifyAuthenticator(
		config.NewSpotifyCredentialStore(p), webServer, spotifyFlow, oauth2util.GenerateOutput{}, "state",
	)

	persistence := new(mocks.AppleMusicAuthPersistenceMock)
	persistence.On("Read").Return(nil, assert.AnError)
	persistence.On("Write", mock.Anything).Return(nil)

	storefront := new(mocks.StorefrontClientMock)
	storefront.On("SetAuth", mock.Anything).Return()
	storefront.On("Storefront", mock.Anything).Return("us", nil)

	logger := new(mocks.LoggerMock)
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()

	appleAuth := applemusic_ucs.NewAuthenticator(persistence, storefront, p, logger)

	registry := music.NewProviderRegistry().
		Register(music.ProviderSpotify, new(mocks.MusicProviderMock), spotifyAuth).
		Register(music.ProviderAppleMusic, new(mocks.MusicProviderMock), appleAuth)

	return &credentialsFixture{
		prompter: p,
		gateway:  NewRootCmdGateway(logger, registry, new(mocks.SetlistFMGetSetlistByIDUseCaseMock)),
	}
}

const (
	spotifyIDPrompt     = "What's your Spotify client ID?"
	spotifySecretPrompt = "What's your Spotify client secret?"
	appleBearerPrompt   = "What's your Apple Music web-player bearer token?"
	appleMUTPrompt      = "What's your Apple Music media-user-token cookie value?"
)

func promptTitles(p *mocks.PrompterMock) []string {
	titles := []string{}
	for _, c := range p.Calls {
		titles = append(titles, c.Arguments.String(0))
	}

	return titles
}

func TestAppleMusicAuthenticationNeverAsksForSpotifyCredentials(t *testing.T) {
	f := newCredentialsFixture(t)
	f.prompter.On("Secret", appleBearerPrompt).Return("bearer", nil)
	f.prompter.On("Secret", appleMUTPrompt).Return("mut", nil)

	assert.NoError(t, f.gateway.Authenticate(context.Background(), music.ProviderAppleMusic))

	assert.Equal(t, []string{appleBearerPrompt, appleMUTPrompt}, promptTitles(f.prompter))
}

func TestSpotifyAuthenticationNeverAsksForAppleMusicCredentials(t *testing.T) {
	f := newCredentialsFixture(t)
	f.prompter.On("Text", spotifyIDPrompt).Return("the-id", nil)
	f.prompter.On("Secret", spotifySecretPrompt).Return("the-secret", nil)

	assert.NoError(t, f.gateway.Authenticate(context.Background(), music.ProviderSpotify))

	assert.Equal(t, []string{spotifyIDPrompt, spotifySecretPrompt}, promptTitles(f.prompter))
}
