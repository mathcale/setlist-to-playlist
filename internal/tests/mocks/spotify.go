package mocks

import (
	"context"
	"net/http"

	"github.com/stretchr/testify/mock"
	"github.com/zmb3/spotify/v2"

	oauth2util "github.com/mathcale/setlist-to-playlist/internal/pkg/oauth2"
)

type SpotifyAuthCallbackUseCaseMock struct {
	mock.Mock
}

func (m *SpotifyAuthCallbackUseCaseMock) Execute(
	ctx context.Context,
	r *http.Request,
	state string,
	genCodes oauth2util.GenerateOutput,
) (*spotify.Client, error) {
	args := m.Called(ctx, r, state, genCodes)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*spotify.Client), args.Error(1)
}
