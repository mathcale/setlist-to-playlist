package spotify

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"

	"github.com/mathcale/setlist-to-playlist/internal/entities/music"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/logger"
	oauth2util "github.com/mathcale/setlist-to-playlist/internal/pkg/oauth2"
)

type SpotifyClientInterface interface {
	GetToken(ctx context.Context, r *http.Request, state string, genCodes oauth2util.GenerateOutput) (*oauth2.Token, error)
	GetAuthURL(state string, genCodes oauth2util.GenerateOutput) string
	NewAPIClient(ctx context.Context, tok *oauth2.Token) *spotify.Client
	SetAuthenticatedClient(ch chan AuthenticatedClient)
	SetAuthenticatedClientFromInstance(client AuthenticatedClient)
	CurrentUser(ctx context.Context) (*spotify.PrivateUser, error)
	CurrentSession() (*oauth2.Token, error)
	RefreshToken(ctx context.Context, tok *oauth2.Token) (*oauth2.Token, error)
	music.MusicProvider
}

const searchLimit = 10

type AuthenticatedClient struct {
	spotify.Client
}

type SpotifyClient struct {
	Auth                *spotifyauth.Authenticator
	AuthenticatedClient AuthenticatedClient
	Logger              logger.LoggerInterface
}

func NewSpotifyClient(
	logger logger.LoggerInterface,
	redirURL string,
	clientID string,
	clientSecret string,
) SpotifyClientInterface {
	return &SpotifyClient{
		Auth: spotifyauth.New(
			spotifyauth.WithRedirectURL(redirURL),
			spotifyauth.WithClientID(clientID),
			spotifyauth.WithClientSecret(clientSecret),
			spotifyauth.WithScopes(spotifyauth.ScopeUserReadEmail, spotifyauth.ScopePlaylistModifyPublic),
		),
		AuthenticatedClient: AuthenticatedClient{},
		Logger:              logger,
	}
}

func (c *SpotifyClient) GetAuthURL(state string, genCodes oauth2util.GenerateOutput) string {
	return c.Auth.AuthURL(
		state,
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("code_challenge", genCodes.CodeChallenge),
	)
}

func (c *SpotifyClient) GetToken(
	ctx context.Context,
	r *http.Request,
	state string,
	genCodes oauth2util.GenerateOutput,
) (*oauth2.Token, error) {
	token, err := c.Auth.Token(
		ctx,
		state,
		r,
		oauth2.SetAuthURLParam("code_verifier", genCodes.CodeVerifier),
	)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (c *SpotifyClient) NewAPIClient(ctx context.Context, tok *oauth2.Token) *spotify.Client {
	return spotify.New(c.Auth.Client(ctx, tok))
}

func (c *SpotifyClient) SetAuthenticatedClient(ch chan AuthenticatedClient) {
	c.AuthenticatedClient = <-ch
}

func (c *SpotifyClient) SetAuthenticatedClientFromInstance(client AuthenticatedClient) {
	c.AuthenticatedClient = client
}

func (c *SpotifyClient) CurrentUser(ctx context.Context) (*spotify.PrivateUser, error) {
	return c.AuthenticatedClient.CurrentUser(ctx)
}

func (c *SpotifyClient) CurrentSession() (*oauth2.Token, error) {
	return c.AuthenticatedClient.Token()
}

func (c *SpotifyClient) RefreshToken(
	ctx context.Context,
	tok *oauth2.Token,
) (*oauth2.Token, error) {
	return c.Auth.RefreshToken(ctx, tok)
}

func (c *SpotifyClient) SearchCandidates(
	ctx context.Context,
	title string,
	artist string,
) ([]music.Candidate, error) {
	q := fmt.Sprintf(`"%s"%%20artist:%s`, strings.ToLower(title), strings.ToLower(artist))

	c.Logger.Debug("Searching for track", map[string]interface{}{
		"query": q,
	})

	res, err := c.AuthenticatedClient.Search(
		ctx,
		q,
		spotify.SearchTypeTrack,
		spotify.Limit(searchLimit),
	)
	if err != nil {
		c.Logger.Error("Failed to search for track", err, map[string]interface{}{
			"query": q,
		})

		return nil, err
	}

	if res.Tracks == nil {
		return nil, nil
	}

	candidates := make([]music.Candidate, 0, len(res.Tracks.Tracks))

	for _, t := range res.Tracks.Tracks {
		artistName := ""
		if len(t.Artists) > 0 {
			artistName = t.Artists[0].Name
		}

		candidates = append(candidates, music.Candidate{
			TrackID:   t.ID.String(),
			Title:     t.Name,
			Artist:    artistName,
			Album:     t.Album.Name,
			AlbumType: mapAlbumType(t.Album.AlbumType),
		})
	}

	return candidates, nil
}

func mapAlbumType(t string) music.AlbumType {
	switch strings.ToLower(t) {
	case "album":
		return music.AlbumTypeAlbum
	case "single":
		return music.AlbumTypeSingle
	case "compilation":
		return music.AlbumTypeCompilation
	default:
		return music.AlbumTypeUnknown
	}
}

func (c *SpotifyClient) CreatePlaylist(
	ctx context.Context,
	title string,
	description string,
) (*music.Playlist, error) {
	user, err := c.CurrentUser(ctx)
	if err != nil {
		return nil, err
	}

	playlist, err := c.AuthenticatedClient.CreatePlaylistForUser(
		ctx,
		user.ID,
		title,
		description,
		true,
		false,
	)
	if err != nil {
		return nil, err
	}

	return &music.Playlist{
		ID:  playlist.ID.String(),
		URL: playlist.ExternalURLs["spotify"],
	}, nil
}

func (c *SpotifyClient) AddTracks(ctx context.Context, playlistID string, trackIDs []string) error {
	c.Logger.Debug("Adding tracks to playlist...", map[string]interface{}{
		"playlist_id": playlistID,
		"song_ids":    trackIDs,
	})

	ids := make([]spotify.ID, len(trackIDs))
	for i, id := range trackIDs {
		ids[i] = spotify.ID(id)
	}

	_, err := c.AuthenticatedClient.AddTracksToPlaylist(ctx, spotify.ID(playlistID), ids...)

	return err
}
