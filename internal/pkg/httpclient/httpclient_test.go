package httpclient_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mathcale/setlist-to-playlist/internal/pkg/httpclient"
)

type payload struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

func newClient(srv *httptest.Server) httpclient.HttpClientInterface {
	return httpclient.NewHttpClient(srv.URL, 2*time.Second)
}

func TestPost_SendsJSONBodyAndDecodesResponse(t *testing.T) {
	var gotMethod, gotPath, gotCT, gotAuth string
	var gotBody []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		gotCT, gotAuth = r.Header.Get("Content-Type"), r.Header.Get("Authorization")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"name":"created","tags":["a"]}`))
	}))
	defer srv.Close()

	var out payload
	err := newClient(srv).Post("/playlists", map[string]interface{}{"Authorization": "Bearer x"},
		payload{Name: "mix", Tags: []string{"x", "y"}}, &out)

	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/playlists", gotPath)
	assert.Equal(t, "application/json", gotCT)
	assert.Equal(t, "Bearer x", gotAuth)
	assert.JSONEq(t, `{"name":"mix","tags":["x","y"]}`, string(gotBody))
	assert.Equal(t, payload{Name: "created", Tags: []string{"a"}}, out)
}

func TestPost_EmptyResponseBody(t *testing.T) {
	for _, code := range []int{http.StatusCreated, http.StatusNoContent} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(code)
		}))

		var out payload
		assert.NoError(t, newClient(srv).Post("/x", nil, payload{}, &out), "code %d with response obj", code)
		assert.NoError(t, newClient(srv).Post("/x", nil, payload{}, nil), "code %d with nil response obj", code)
		srv.Close()
	}
}

func TestStatusErrors_GetAndPost(t *testing.T) {
	for _, code := range []int{401, 403, 500} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(code)
			_, _ = w.Write([]byte(`{"error":"nope"}`))
		}))
		c := newClient(srv)

		calls := map[string]error{
			"GET":  c.Get("/x", nil, &payload{}),
			"POST": c.Post("/x", nil, payload{}, &payload{}),
		}

		for name, err := range calls {
			var se *httpclient.StatusError
			require.True(t, errors.As(err, &se), "%s %d: %v", name, code, err)
			assert.Equal(t, code, se.StatusCode, "%s", name)
		}
		srv.Close()
	}
}

func TestGet_DecodesResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		_ = json.NewEncoder(w).Encode(payload{Name: "ok"})
	}))
	defer srv.Close()

	var out payload
	require.NoError(t, newClient(srv).Get("/x", nil, &out))
	assert.Equal(t, "ok", out.Name)
}
