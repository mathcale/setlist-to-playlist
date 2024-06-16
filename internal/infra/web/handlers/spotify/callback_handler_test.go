package spotify

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/zmb3/spotify/v2"

	client "github.com/mathcale/setlist-to-playlist/internal/clients/spotify"
	oauth2util "github.com/mathcale/setlist-to-playlist/internal/pkg/oauth2"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/responsehandler"
	"github.com/mathcale/setlist-to-playlist/internal/tests/mocks"
)

type SpotifyAuthCallbackWebHandlerTestSuite struct {
	suite.Suite
	CallbackUseCaseMock           *mocks.SpotifyAuthCallbackUseCaseMock
	ResponseHandler               *responsehandler.WebResponseHandler
	GenCodes                      oauth2util.GenerateOutput
	State                         string
	SpotifyAuthCallbackWebHandler SpotifyAuthCallbackWebHandlerInterface
	Channel                       chan client.AuthenticatedClient
}

func TestSpotifyAuthCallbackWebHandler(t *testing.T) {
	suite.Run(t, new(SpotifyAuthCallbackWebHandlerTestSuite))
}

func (s *SpotifyAuthCallbackWebHandlerTestSuite) SetupTest() {
	s.CallbackUseCaseMock = new(mocks.SpotifyAuthCallbackUseCaseMock)
	s.ResponseHandler = &responsehandler.WebResponseHandler{}
	s.GenCodes = oauth2util.GenerateOutput{
		CodeVerifier:  "any-code-verifier",
		CodeChallenge: "code-challenge",
	}
	s.State = "any-state"
	s.Channel = make(chan client.AuthenticatedClient)

	s.SpotifyAuthCallbackWebHandler = NewSpotifyAuthCallbackWebHandler(
		s.CallbackUseCaseMock,
		s.ResponseHandler,
		s.GenCodes,
		s.State,
		s.Channel,
	)
}

func (s *SpotifyAuthCallbackWebHandlerTestSuite) cleanMocks() {
	s.CallbackUseCaseMock.ExpectedCalls = nil
	s.CallbackUseCaseMock.Calls = nil
}

func (s *SpotifyAuthCallbackWebHandlerTestSuite) TestHandle() {
	// FIXME: getting stuck in the channel
	s.Run("should handle callback", func() {
		r := httptest.NewRequest(http.MethodGet, "/callback", nil)
		w := httptest.NewRecorder()

		s.CallbackUseCaseMock.On("Execute", r.Context(), r, s.State, s.GenCodes).Return(&spotify.Client{}, nil)

		s.SpotifyAuthCallbackWebHandler.Handle(w, r)

		res := w.Result()
		defer res.Body.Close()

		expected := `{"message":"Spotify login completed, you can close this page now."}`

		s.Equal(http.StatusOK, res.StatusCode)
		s.Equal(expected, strings.TrimSuffix(w.Body.String(), "\n"))
	})
}
