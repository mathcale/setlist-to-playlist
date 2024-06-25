package spotify

import (
	"embed"
	"html/template"
	"net/http"

	client "github.com/mathcale/setlist-to-playlist/internal/clients/spotify"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/logger"
	oauth2util "github.com/mathcale/setlist-to-playlist/internal/pkg/oauth2"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/responsehandler"
	uc "github.com/mathcale/setlist-to-playlist/internal/usecases/spotify"
)

type SpotifyAuthCallbackWebHandlerInterface interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

type SpotifyAuthCallbackWebHandler struct {
	Logger               logger.LoggerInterface
	CallbackUseCase      uc.SpotifyAuthCallbackUseCaseInterface
	ResponseHandler      responsehandler.WebResponseHandlerInterface
	GeneratedPKCECodes   oauth2util.GenerateOutput
	State                string
	SpotifyClientChannel chan client.AuthenticatedClient
}

//go:embed static/callback.html
var callbackHTML embed.FS

func NewSpotifyAuthCallbackWebHandler(
	l logger.LoggerInterface,
	uc uc.SpotifyAuthCallbackUseCaseInterface,
	rh responsehandler.WebResponseHandlerInterface,
	genCodes oauth2util.GenerateOutput,
	state string,
	ch chan client.AuthenticatedClient,
) SpotifyAuthCallbackWebHandlerInterface {
	return &SpotifyAuthCallbackWebHandler{
		Logger:               l,
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

	h.renderCallbackPage(w)

	authClient := client.AuthenticatedClient{
		Client: *cl,
	}

	h.SpotifyClientChannel <- authClient
}

func (h *SpotifyAuthCallbackWebHandler) renderCallbackPage(w http.ResponseWriter) {
	t, _ := template.ParseFS(callbackHTML, "static/callback.html")

	w.Header().Add("Content-Type", "text/html")

	if err := t.Execute(w, nil); err != nil {
		h.Logger.Error("Error rendering callback page", err, nil)
	}
}
