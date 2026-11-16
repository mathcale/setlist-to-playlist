package applemusic

import (
	"context"
	"errors"
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
// authenticator through SetAuth. It implements music.MusicProvider; search
// and playlist creation arrive with ATMT-168 and return
// music.ErrNotImplemented until then.
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
		return "", err
	}

	if len(resp.Data) == 0 || resp.Data[0].ID == "" {
		return "", errors.New("apple music returned no storefront")
	}

	c.storefront = resp.Data[0].ID

	return c.storefront, nil
}

func (c *Client) SearchCandidates(context.Context, string, string) ([]music.Candidate, error) {
	return nil, music.ErrNotImplemented
}

func (c *Client) CreatePlaylist(context.Context, string, string) (*music.Playlist, error) {
	return nil, music.ErrNotImplemented
}

func (c *Client) AddTracks(context.Context, string, []string) error {
	return music.ErrNotImplemented
}
