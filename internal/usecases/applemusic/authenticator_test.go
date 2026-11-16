package applemusic

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mathcale/setlist-to-playlist/internal/entities/applemusic"
	"github.com/mathcale/setlist-to-playlist/internal/tests/mocks"
)

type fixture struct {
	persistence *mocks.AppleMusicAuthPersistenceMock
	client      *mocks.StorefrontClientMock
	prompter    *mocks.PrompterMock
	logger      *mocks.LoggerMock
	auth        *Authenticator
}

func newFixture() *fixture {
	f := &fixture{
		persistence: new(mocks.AppleMusicAuthPersistenceMock),
		client:      new(mocks.StorefrontClientMock),
		prompter:    new(mocks.PrompterMock),
		logger:      new(mocks.LoggerMock),
	}
	f.auth = NewAuthenticator(f.persistence, f.client, f.prompter, f.logger)

	return f
}

var tokens = applemusic.AuthData{BearerToken: "bearer", MediaUserToken: "mut"}

func TestAuthenticate(t *testing.T) {
	t.Run("first use prompts for both secrets, saves them and logs the storefront", func(t *testing.T) {
		f := newFixture()
		f.persistence.On("Read").Return(&applemusic.AuthData{}, nil)
		f.prompter.On("Secret", mock.Anything).Return("bearer", nil).Once()
		f.prompter.On("Secret", mock.Anything).Return(" mut ", nil).Once()
		f.persistence.On("Write", tokens).Return(nil)
		f.client.On("SetAuth", tokens).Return()
		f.client.On("Storefront", mock.Anything).Return("us", nil)
		f.logger.On("Info", "Authenticated on Apple Music (storefront: us)", mock.Anything).Return()

		assert.NoError(t, f.auth.Authenticate(context.Background()))

		f.prompter.AssertNumberOfCalls(t, "Secret", 2)
		f.persistence.AssertCalled(t, "Write", tokens)
		f.logger.AssertExpectations(t)
	})

	t.Run("saved tokens are used without prompting", func(t *testing.T) {
		f := newFixture()
		saved := tokens
		f.persistence.On("Read").Return(&saved, nil)
		f.client.On("SetAuth", tokens).Return()
		f.client.On("Storefront", mock.Anything).Return("us", nil)
		f.logger.On("Info", mock.Anything, mock.Anything).Return()

		assert.NoError(t, f.auth.Authenticate(context.Background()))

		f.prompter.AssertNotCalled(t, "Secret", mock.Anything)
		f.persistence.AssertNotCalled(t, "Write", mock.Anything)
	})

	t.Run("rejects empty input and saves nothing", func(t *testing.T) {
		f := newFixture()
		f.persistence.On("Read").Return(nil, errors.New("missing"))
		f.prompter.On("Secret", mock.Anything).Return("  ", nil)

		assert.Error(t, f.auth.Authenticate(context.Background()))

		f.persistence.AssertNotCalled(t, "Write", mock.Anything)
		f.client.AssertNotCalled(t, "Storefront", mock.Anything)
	})

	t.Run("returns the storefront error", func(t *testing.T) {
		f := newFixture()
		saved := tokens
		f.persistence.On("Read").Return(&saved, nil)
		f.client.On("SetAuth", tokens).Return()
		f.client.On("Storefront", mock.Anything).Return("", errors.New("401"))

		assert.EqualError(t, f.auth.Authenticate(context.Background()), "401")
	})
}
