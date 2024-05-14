package responsehandler

import (
	"encoding/json"
	"net/http"
)

type WebResponseHandlerInterface interface {
	Respond(w http.ResponseWriter, statusCode int, data interface{})
	RespondWithError(w http.ResponseWriter, statusCode int, err error)
}

type WebResponseHandler struct{}

func NewWebResponseHandler() *WebResponseHandler {
	return &WebResponseHandler{}
}

func (h *WebResponseHandler) Respond(w http.ResponseWriter, statusCode int, data interface{}) {
	setHeaders(w)
	w.WriteHeader(statusCode)

	if data != nil {
		json.NewEncoder(w).Encode(&data)
	}
}

func (h *WebResponseHandler) RespondWithError(w http.ResponseWriter, statusCode int, err error) {
	setHeaders(w)
	w.WriteHeader(statusCode)

	json.NewEncoder(w).Encode(map[string]string{
		"message": err.Error(),
	})
}

func setHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Accept", "application/json")
}
