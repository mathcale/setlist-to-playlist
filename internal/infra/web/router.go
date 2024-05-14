package web

import (
	spotify_handlers "github.com/mathcale/setlist-to-playlist/internal/infra/web/handlers/spotify"
)

type WebRouterInterface interface {
	Build() []RouteHandler
}

type WebRouter struct {
	SpotifyAuthCallbackWebHandler spotify_handlers.SpotifyAuthCallbackWebHandlerInterface
}

func NewWebRouter(
	sacwh spotify_handlers.SpotifyAuthCallbackWebHandlerInterface,
) *WebRouter {
	return &WebRouter{
		SpotifyAuthCallbackWebHandler: sacwh,
	}
}

func (wr *WebRouter) Build() []RouteHandler {
	return []RouteHandler{
		{
			Path:        "/callback",
			Method:      "GET",
			HandlerFunc: wr.SpotifyAuthCallbackWebHandler.Handle,
		},
	}
}
