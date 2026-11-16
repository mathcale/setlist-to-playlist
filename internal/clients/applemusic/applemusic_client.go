package applemusic

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/mathcale/setlist-to-playlist/internal/entities/applemusic"
	"github.com/mathcale/setlist-to-playlist/internal/entities/music"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/httpclient"
)

const (
	BaseURL        = "https://api.music.apple.com"
	Origin         = "https://music.apple.com"
	StorefrontPath = "/v1/me/storefront"
)

type storefrontResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

// Client talks to the Apple Music web API. Credentials are supplied by the
// authenticator through SetAuth. It implements music.MusicProvider.
type Client struct {
	httpClient httpclient.HttpClientInterface

	mu         sync.Mutex
	auth       applemusic.AuthData
	storefront string
}

var _ music.MusicProvider = (*Client)(nil)

func NewClient(httpClient httpclient.HttpClientInterface) *Client {
	return &Client{httpClient: httpClient}
}

// SetAuth sets the tokens used on every request and forgets any cached
// storefront (it belongs to the previous tokens).
func (c *Client) SetAuth(auth applemusic.AuthData) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.auth = auth
	c.storefront = ""
}

func (c *Client) headers() map[string]interface{} {
	return map[string]interface{}{
		"Authorization":    "Bearer " + c.auth.BearerToken,
		"Music-User-Token": c.auth.MediaUserToken,
		"Origin":           Origin,
	}
}

// mapErr turns 401/403 responses into music.ErrCredentialsExpired; any other
// error is returned untouched.
func mapErr(err error) error {
	var se *httpclient.StatusError
	if errors.As(err, &se) && (se.StatusCode == http.StatusUnauthorized || se.StatusCode == http.StatusForbidden) {
		return fmt.Errorf("%w: apple music responded with HTTP %d", music.ErrCredentialsExpired, se.StatusCode)
	}

	return err
}

// Storefront returns the user's storefront (e.g. "us"), fetching it on the
// first call and reusing it for the rest of the run.
func (c *Client) Storefront(_ context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.storefront != "" {
		return c.storefront, nil
	}

	var resp storefrontResponse
	if err := c.httpClient.Get(StorefrontPath, c.headers(), &resp); err != nil {
		return "", mapErr(err)
	}

	if len(resp.Data) == 0 || resp.Data[0].ID == "" {
		return "", errors.New("apple music returned no storefront")
	}

	c.storefront = resp.Data[0].ID

	return c.storefront, nil
}

type searchResponse struct {
	Results struct {
		Songs *struct {
			Data []songResource `json:"data"`
		} `json:"songs"`
	} `json:"results"`
}

type songResource struct {
	ID         string `json:"id"`
	Attributes struct {
		Name       string `json:"name"`
		ArtistName string `json:"artistName"`
		AlbumName  string `json:"albumName"`
	} `json:"attributes"`
	Relationships struct {
		Albums struct {
			Data []struct {
				Attributes struct {
					Name          string `json:"name"`
					IsSingle      bool   `json:"isSingle"`
					IsCompilation bool   `json:"isCompilation"`
				} `json:"attributes"`
			} `json:"data"`
		} `json:"albums"`
	} `json:"relationships"`
}

// albumType maps Apple's album attributes to the shared AlbumType. Apple has
// no explicit EP flag, so EPs are detected by the " - EP" name suffix.
func (s songResource) albumType() music.AlbumType {
	albums := s.Relationships.Albums.Data
	if len(albums) == 0 {
		return music.AlbumTypeUnknown
	}

	a := albums[0].Attributes

	switch {
	case a.IsCompilation:
		return music.AlbumTypeCompilation
	case a.IsSingle:
		return music.AlbumTypeSingle
	case strings.HasSuffix(a.Name, " - EP"):
		return music.AlbumTypeEP
	default:
		return music.AlbumTypeAlbum
	}
}

// SearchCandidates searches the catalog for songs. The song's album is
// requested with include[songs]=albums (Apple's documented way to embed a
// relationship in search results) and fields[albums] trims it to the
// attributes needed to classify the album type.
func (c *Client) SearchCandidates(ctx context.Context, title string, artist string) ([]music.Candidate, error) {
	storefront, err := c.Storefront(ctx)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Set("types", "songs")
	q.Set("term", title+" "+artist)
	q.Set("limit", "10")
	q.Set("include[songs]", "albums")
	q.Set("fields[albums]", "name,isSingle,isCompilation")

	var resp searchResponse
	if err := c.httpClient.Get(
		"/v1/catalog/"+url.PathEscape(storefront)+"/search?"+q.Encode(), c.headers(), &resp,
	); err != nil {
		return nil, mapErr(err)
	}

	candidates := []music.Candidate{}
	if resp.Results.Songs == nil {
		return candidates, nil
	}

	for _, s := range resp.Results.Songs.Data {
		candidates = append(candidates, music.Candidate{
			TrackID:   s.ID,
			Title:     s.Attributes.Name,
			Artist:    s.Attributes.ArtistName,
			Album:     s.Attributes.AlbumName,
			AlbumType: s.albumType(),
		})
	}

	return candidates, nil
}

type createPlaylistRequest struct {
	Attributes struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"attributes"`
}

type createPlaylistResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

// PlaylistURLPrefix is a library link: it only opens for the owning account.
const PlaylistURLPrefix = "https://music.apple.com/library/playlist/"

func (c *Client) CreatePlaylist(_ context.Context, title string, description string) (*music.Playlist, error) {
	var body createPlaylistRequest
	body.Attributes.Name = title
	body.Attributes.Description = description

	var resp createPlaylistResponse
	if err := c.httpClient.Post("/v1/me/library/playlists", c.headers(), body, &resp); err != nil {
		return nil, mapErr(err)
	}

	if len(resp.Data) == 0 || resp.Data[0].ID == "" {
		return nil, errors.New("apple music returned no playlist id")
	}

	return &music.Playlist{ID: resp.Data[0].ID, URL: PlaylistURLPrefix + resp.Data[0].ID}, nil
}

type trackRef struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type addTracksRequest struct {
	Data []trackRef `json:"data"`
}

func (c *Client) AddTracks(_ context.Context, playlistID string, trackIDs []string) error {
	body := addTracksRequest{Data: make([]trackRef, 0, len(trackIDs))}
	for _, id := range trackIDs {
		body.Data = append(body.Data, trackRef{ID: id, Type: "songs"})
	}

	return mapErr(c.httpClient.Post("/v1/me/library/playlists/"+url.PathEscape(playlistID)+"/tracks", c.headers(), body, nil))
}
