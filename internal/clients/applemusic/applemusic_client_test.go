package applemusic

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mathcale/setlist-to-playlist/internal/entities/applemusic"
	"github.com/mathcale/setlist-to-playlist/internal/entities/music"
	"github.com/mathcale/setlist-to-playlist/internal/tests/mocks"
)

func newClient(h *mocks.HttpClientMock) *Client {
	c := NewClient(h)
	c.SetAuth(applemusic.AuthData{BearerToken: "bearer", MediaUserToken: "mut"})
	return c
}

func TestStorefront(t *testing.T) {
	wantHeaders := map[string]interface{}{
		"Authorization":    "Bearer bearer",
		"Music-User-Token": "mut",
		"Origin":           "https://music.apple.com",
	}

	t.Run("fetches the storefront with the required headers and caches it", func(t *testing.T) {
		h := new(mocks.HttpClientMock)
		h.On("Get", "/v1/me/storefront", wantHeaders, mock.Anything).
			Run(func(args mock.Arguments) {
				resp := args.Get(2).(*storefrontResponse)
				resp.Data = append(resp.Data, struct {
					ID string `json:"id"`
				}{ID: "br"})
			}).Return(nil).Once()

		c := newClient(h)

		for i := 0; i < 2; i++ {
			sf, err := c.Storefront(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, "br", sf)
		}

		h.AssertNumberOfCalls(t, "Get", 1)
	})

	t.Run("returns the http client error", func(t *testing.T) {
		h := new(mocks.HttpClientMock)
		h.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("boom"))

		_, err := newClient(h).Storefront(context.Background())
		assert.EqualError(t, err, "boom")
	})

	t.Run("errors when the response has no storefront", func(t *testing.T) {
		h := new(mocks.HttpClientMock)
		h.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		_, err := newClient(h).Storefront(context.Background())
		assert.Error(t, err)
	})
}

func TestProviderMethodsAreNotImplementedYet(t *testing.T) {
	c := newClient(new(mocks.HttpClientMock))

	_, err := c.SearchCandidates(context.Background(), "t", "a")
	assert.ErrorIs(t, err, music.ErrNotImplemented)
	_, err = c.CreatePlaylist(context.Background(), "t", "d")
	assert.ErrorIs(t, err, music.ErrNotImplemented)
	assert.ErrorIs(t, c.AddTracks(context.Background(), "p", nil), music.ErrNotImplemented)
}
