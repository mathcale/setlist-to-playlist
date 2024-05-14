package web

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/mathcale/setlist-to-playlist/internal/pkg/logger"
)

type WebServerInterface interface {
	Start()
}

type RouteHandler struct {
	Path        string
	Method      string
	HandlerFunc http.HandlerFunc
}

type WebServer struct {
	Router        chi.Router
	Handlers      []RouteHandler
	WebServerPort int64
	Logger        logger.LoggerInterface
}

func NewWebServer(
	serverPort int64,
	logger logger.LoggerInterface,
	handlers []RouteHandler,
) WebServerInterface {
	return &WebServer{
		Router:        chi.NewRouter(),
		Handlers:      handlers,
		WebServerPort: serverPort,
		Logger:        logger,
	}
}

func (s *WebServer) Start() {
	s.Router.Use(s.Logger.NewChiServerLogger())

	for _, h := range s.Handlers {
		s.Logger.Debug(fmt.Sprintf("Registering route %s %s", h.Method, h.Path), nil)
		s.Router.MethodFunc(h.Method, h.Path, h.HandlerFunc)
	}

	s.Logger.Debug(fmt.Sprintf("Starting webserver on port [%d]", s.WebServerPort), nil)

	go http.ListenAndServe(fmt.Sprintf(":%d", s.WebServerPort), s.Router)
}
