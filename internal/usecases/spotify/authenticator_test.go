package spotify

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	oauth2util "github.com/mathcale/setlist-to-playlist/internal/pkg/oauth2"
	"github.com/mathcale/setlist-to-playlist/internal/tests/mocks"
)

func TestSpotifyAuthenticator_EnsuresCredentialsThenStartsServerThenAuthenticates(t *testing.T) {
	var order []string
	codes := oauth2util.GenerateOutput{CodeChallenge: "c", CodeVerifier: "v"}

	creds := new(mocks.CredentialsEnsurerMock)
	ws := new(mocks.WebServerMock)
	uc := new(mocks.SpotifyUserAuthenticationUseCaseMock)

	creds.On("Ensure").Run(func(mock.Arguments) { order = append(order, "ensure") }).Return(nil)
	ws.On("Start").Run(func(mock.Arguments) { order = append(order, "server") }).Return()
	uc.On("Execute", mock.Anything, codes, "state").
		Run(func(mock.Arguments) { order = append(order, "auth") }).Return(nil)

	a := NewSpotifyAuthenticator(creds, ws, uc, codes, "state")

	assert.NoError(t, a.Authenticate(context.Background()))
	assert.Equal(t, []string{"ensure", "server", "auth"}, order)
	uc.AssertExpectations(t)
}

func TestSpotifyAuthenticator_StopsWhenCredentialsFail(t *testing.T) {
	creds := new(mocks.CredentialsEnsurerMock)
	ws := new(mocks.WebServerMock)
	uc := new(mocks.SpotifyUserAuthenticationUseCaseMock)
	creds.On("Ensure").Return(errors.New("aborted"))

	a := NewSpotifyAuthenticator(creds, ws, uc, oauth2util.GenerateOutput{}, "s")

	assert.EqualError(t, a.Authenticate(context.Background()), "aborted")
	ws.AssertNotCalled(t, "Start")
	uc.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything, mock.Anything)
}

func TestSpotifyAuthenticator_PropagatesAuthErrorAndStartsServerOnce(t *testing.T) {
	creds := new(mocks.CredentialsEnsurerMock)
	ws := new(mocks.WebServerMock)
	uc := new(mocks.SpotifyUserAuthenticationUseCaseMock)
	creds.On("Ensure").Return(nil)
	ws.On("Start").Return()
	uc.On("Execute", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("denied"))

	a := NewSpotifyAuthenticator(creds, ws, uc, oauth2util.GenerateOutput{}, "s")

	assert.EqualError(t, a.Authenticate(context.Background()), "denied")
	assert.EqualError(t, a.Authenticate(context.Background()), "denied")
	ws.AssertNumberOfCalls(t, "Start", 1)
}
