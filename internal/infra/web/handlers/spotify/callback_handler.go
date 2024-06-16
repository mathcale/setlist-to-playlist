package spotify

import (
	"net/http"

	client "github.com/mathcale/setlist-to-playlist/internal/clients/spotify"
	oauth2util "github.com/mathcale/setlist-to-playlist/internal/pkg/oauth2"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/responsehandler"
	uc "github.com/mathcale/setlist-to-playlist/internal/usecases/spotify"
)

type SpotifyAuthCallbackWebHandlerInterface interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

type SpotifyAuthCallbackWebHandler struct {
	CallbackUseCase      uc.SpotifyAuthCallbackUseCaseInterface
	ResponseHandler      responsehandler.WebResponseHandlerInterface
	GeneratedPKCECodes   oauth2util.GenerateOutput
	State                string
	SpotifyClientChannel chan client.AuthenticatedClient
}

func NewSpotifyAuthCallbackWebHandler(
	uc uc.SpotifyAuthCallbackUseCaseInterface,
	rh responsehandler.WebResponseHandlerInterface,
	genCodes oauth2util.GenerateOutput,
	state string,
	ch chan client.AuthenticatedClient,
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
	cl, err := h.CallbackUseCase.Execute(r.Context(), r, h.State, h.GeneratedPKCECodes)
	if err != nil {
		h.ResponseHandler.RespondWithError(w, http.StatusInternalServerError, err)
		return
	}

	h.ResponseHandler.Respond(w, http.StatusOK, map[string]string{
		"message": "Spotify login completed, you can close this page now.",
	})

	authClient := client.AuthenticatedClient{
		Client: *cl,
	}

	h.SpotifyClientChannel <- authClient
}
