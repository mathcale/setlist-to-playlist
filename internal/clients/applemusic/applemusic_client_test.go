package applemusic

import (
	"context"
	"encoding/json"
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

var wantHeaders = map[string]interface{}{
	"Authorization":    "Bearer bearer",
	"Music-User-Token": "mut",
	"Origin":           "https://music.apple.com",
}

const searchEndpoint = "/v1/catalog/br/search?fields%5Balbums%5D=name%2CisSingle%2CisCompilation" +
	"&include%5Bsongs%5D=albums&limit=10&term=Song+Artist&types=songs"

func withStorefront(h *mocks.HttpClientMock) {
	h.On("Get", "/v1/me/storefront", wantHeaders, mock.Anything).
		Run(func(args mock.Arguments) {
			resp := args.Get(2).(*storefrontResponse)
			resp.Data = append(resp.Data, struct {
				ID string `json:"id"`
			}{ID: "br"})
		}).Return(nil)
}

func decodeInto(t *testing.T, raw string) func(mock.Arguments) {
	return func(args mock.Arguments) {
		assert.NoError(t, json.Unmarshal([]byte(raw), args.Get(len(args)-1)))
	}
}

func TestSearchCandidates(t *testing.T) {
	t.Run("parses songs and maps every album type", func(t *testing.T) {
		h := new(mocks.HttpClientMock)
		withStorefront(h)
		h.On("Get", searchEndpoint, wantHeaders, mock.Anything).
			Run(decodeInto(t, `{"results":{"songs":{"data":[
			 {"id":"1","attributes":{"name":"Song","artistName":"Artist","albumName":"Alb"},
			  "relationships":{"albums":{"data":[{"attributes":{"name":"Alb"}}]}}},
			 {"id":"2","attributes":{"name":"Song","artistName":"Artist","albumName":"Best Of"},
			  "relationships":{"albums":{"data":[{"attributes":{"name":"Best Of","isCompilation":true}}]}}},
			 {"id":"3","attributes":{"name":"Song","artistName":"Artist","albumName":"Song - Single"},
			  "relationships":{"albums":{"data":[{"attributes":{"name":"Song - Single","isSingle":true}}]}}},
			 {"id":"4","attributes":{"name":"Song","artistName":"Artist","albumName":"Thing - EP"},
			  "relationships":{"albums":{"data":[{"attributes":{"name":"Thing - EP"}}]}}},
			 {"id":"5","attributes":{"name":"Song","artistName":"Artist","albumName":"X"}}
			]}}}`)).Return(nil)

		got, err := newClient(h).SearchCandidates(context.Background(), "Song", "Artist")

		assert.NoError(t, err)
		assert.Equal(t, []music.Candidate{
			{TrackID: "1", Title: "Song", Artist: "Artist", Album: "Alb", AlbumType: music.AlbumTypeAlbum},
			{TrackID: "2", Title: "Song", Artist: "Artist", Album: "Best Of", AlbumType: music.AlbumTypeCompilation},
			{TrackID: "3", Title: "Song", Artist: "Artist", Album: "Song - Single", AlbumType: music.AlbumTypeSingle},
			{TrackID: "4", Title: "Song", Artist: "Artist", Album: "Thing - EP", AlbumType: music.AlbumTypeEP},
			{TrackID: "5", Title: "Song", Artist: "Artist", Album: "X", AlbumType: music.AlbumTypeUnknown},
		}, got)
	})

	t.Run("returns an empty slice when there are no songs", func(t *testing.T) {
		h := new(mocks.HttpClientMock)
		withStorefront(h)
		h.On("Get", searchEndpoint, wantHeaders, mock.Anything).Run(decodeInto(t, `{"results":{}}`)).Return(nil)

		got, err := newClient(h).SearchCandidates(context.Background(), "Song", "Artist")

		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Empty(t, got)
	})

	t.Run("returns the http client error", func(t *testing.T) {
		h := new(mocks.HttpClientMock)
		withStorefront(h)
		h.On("Get", searchEndpoint, wantHeaders, mock.Anything).Return(errors.New("boom"))

		_, err := newClient(h).SearchCandidates(context.Background(), "Song", "Artist")
		assert.EqualError(t, err, "boom")
	})

	t.Run("shared ranking picks the studio album over live and compilation", func(t *testing.T) {
		h := new(mocks.HttpClientMock)
		withStorefront(h)
		h.On("Get", searchEndpoint, wantHeaders, mock.Anything).
			Run(decodeInto(t, `{"results":{"songs":{"data":[
			 {"id":"live","attributes":{"name":"Song (Live)","artistName":"Artist","albumName":"Live At X"},
			  "relationships":{"albums":{"data":[{"attributes":{"name":"Live At X"}}]}}},
			 {"id":"hits","attributes":{"name":"Song","artistName":"Artist","albumName":"Hits"},
			  "relationships":{"albums":{"data":[{"attributes":{"name":"Hits","isCompilation":true}}]}}},
			 {"id":"studio","attributes":{"name":"Song","artistName":"Artist","albumName":"Studio"},
			  "relationships":{"albums":{"data":[{"attributes":{"name":"Studio"}}]}}}
			]}}}`)).Return(nil)

		got, err := newClient(h).SearchCandidates(context.Background(), "Song", "Artist")
		assert.NoError(t, err)

		best, ok := music.PickBestCandidate(got, "Artist")
		assert.True(t, ok)
		assert.Equal(t, "studio", best.TrackID)
	})
}

func TestCreatePlaylist(t *testing.T) {
	t.Run("posts the title and description and returns the library url", func(t *testing.T) {
		h := new(mocks.HttpClientMock)
		h.On("Post", "/v1/me/library/playlists", wantHeaders, mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				b, _ := json.Marshal(args.Get(2))
				assert.JSONEq(t, `{"attributes":{"name":"My Setlist","description":"desc"}}`, string(b))
				decodeInto(t, `{"data":[{"id":"p.abc"}]}`)(args)
			}).Return(nil)

		got, err := newClient(h).CreatePlaylist(context.Background(), "My Setlist", "desc")

		assert.NoError(t, err)
		assert.Equal(t, "p.abc", got.ID)
		assert.Equal(t, "https://music.apple.com/library/playlist/p.abc", got.URL)
	})

	t.Run("errors when no playlist id is returned", func(t *testing.T) {
		h := new(mocks.HttpClientMock)
		h.On("Post", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		_, err := newClient(h).CreatePlaylist(context.Background(), "t", "d")
		assert.Error(t, err)
	})

	t.Run("returns the http client error", func(t *testing.T) {
		h := new(mocks.HttpClientMock)
		h.On("Post", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("boom"))

		_, err := newClient(h).CreatePlaylist(context.Background(), "t", "d")
		assert.EqualError(t, err, "boom")
	})
}

func TestAddTracks(t *testing.T) {
	t.Run("posts the tracks in order", func(t *testing.T) {
		h := new(mocks.HttpClientMock)
		h.On("Post", "/v1/me/library/playlists/p.abc/tracks", wantHeaders, mock.Anything, nil).
			Run(func(args mock.Arguments) {
				b, _ := json.Marshal(args.Get(2))
				assert.JSONEq(t, `{"data":[{"id":"3","type":"songs"},{"id":"1","type":"songs"},{"id":"2","type":"songs"}]}`, string(b))
			}).Return(nil)

		assert.NoError(t, newClient(h).AddTracks(context.Background(), "p.abc", []string{"3", "1", "2"}))
		h.AssertNumberOfCalls(t, "Post", 1)
	})

	t.Run("returns the http client error", func(t *testing.T) {
		h := new(mocks.HttpClientMock)
		h.On("Post", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("boom"))

		assert.EqualError(t, newClient(h).AddTracks(context.Background(), "p", []string{"1"}), "boom")
	})
}
