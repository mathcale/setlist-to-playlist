package spotify

import (
	"net/http"

	"github.com/zmb3/spotify/v2"

	oauth2util "github.com/mathcale/setlist-to-playlist/internal/pkg/oauth2"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/responsehandler"
	spotifyuc "github.com/mathcale/setlist-to-playlist/internal/usecases/spotify"
)

type SpotifyAuthCallbackWebHandlerInterface interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

type SpotifyAuthCallbackWebHandler struct {
	CallbackUseCase      spotifyuc.SpotifyAuthCallbackUseCaseInterface
	ResponseHandler      responsehandler.WebResponseHandlerInterface
	GeneratedPKCECodes   oauth2util.GenerateOutput
	State                string
	SpotifyClientChannel chan *spotify.Client
}

func NewSpotifyAuthCallbackWebHandler(
	uc spotifyuc.SpotifyAuthCallbackUseCaseInterface,
	rh responsehandler.WebResponseHandlerInterface,
	genCodes oauth2util.GenerateOutput,
	state string,
	ch chan *spotify.Client,
) SpotifyAuthCallbackWebHandlerInterface {
	return &SpotifyAuthCallbackWebHandler{
		CallbackUseCase:      uc,
		ResponseHandler:      rh,
		GeneratedPKCECodes:   genCodes,
		State:                state,
		SpotifyClientChannel: ch,
	}
}

func (h *SpotifyAuthCallbackWebHandler) Handle(w http.ResponseWriter, r *http.Request) {
	client, err := h.CallbackUseCase.Execute(r.Context(), r, h.State, h.GeneratedPKCECodes)
	if err != nil {
		h.ResponseHandler.RespondWithError(w, http.StatusInternalServerError, err)
		return
	}

	h.SpotifyClientChannel <- client

	h.ResponseHandler.Respond(w, http.StatusOK, map[string]string{
		"message": "Spotify login completed, you can close this page now.",
	})
}
