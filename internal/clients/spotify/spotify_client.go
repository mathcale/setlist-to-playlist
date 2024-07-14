package spotify

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
	"golang.org/x/sync/semaphore"

	entities "github.com/mathcale/setlist-to-playlist/internal/entities/spotify"
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
	FindAllSongsByName(ctx context.Context, name []string, artist string) (*entities.FindAllSongsOutput, error)
	CreatePlaylist(ctx context.Context, title string, description string) (*entities.CreatePlaylistOutput, error)
	AddTracksToPlaylist(ctx context.Context, input entities.AddTracksToPlaylistClientInput) error
}

type AuthenticatedClient struct {
	spotify.Client
}

type SpotifyClient struct {
	Auth                *spotifyauth.Authenticator
	AuthenticatedClient AuthenticatedClient
	Logger              logger.LoggerInterface
}

type FindSongJob struct {
	Songs []string
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

func (c *SpotifyClient) FindAllSongsByName(
	ctx context.Context,
	songNames []string,
	artist string,
) (*entities.FindAllSongsOutput, error) {
	result := &entities.FindAllSongsOutput{
		Artist: artist,
	}

	maxConcurrency := int64(5)
	sem := semaphore.NewWeighted(maxConcurrency)

	for _, n := range songNames {
		if err := sem.Acquire(ctx, 1); err != nil {
			c.Logger.Error("Failed to acquire semaphore", err, nil)
			continue
		}

		go func(song string) {
			defer sem.Release(1)

			q := fmt.Sprintf(`"%s"%%20artist:"%s"`, strings.ToLower(n), strings.ToLower(artist))

			c.Logger.Debug("Searching for track", map[string]interface{}{
				"query": q,
			})

			res, err := c.AuthenticatedClient.Search(
				ctx,
				q,
				spotify.SearchTypeTrack,
				spotify.Limit(1),
			)
			if err != nil {
				c.Logger.Error("Failed to search for track", err, map[string]interface{}{
					"query": q,
				})
			}

			if res.Tracks != nil {
				song := entities.Song{
					ID:    res.Tracks.Tracks[0].ID.String(),
					Title: res.Tracks.Tracks[0].Name,
					Album: res.Tracks.Tracks[0].Album.Name,
				}

				c.Logger.Debug("Found track", map[string]interface{}{
					"id":    song.ID,
					"track": song.Title,
					"album": song.Album,
				})

				result.AddSong(song)
			}
		}(n)
	}

	if err := sem.Acquire(ctx, maxConcurrency); err != nil {
		c.Logger.Error("Failed to acquire semaphore while waiting", err, nil)
	}

	return result, nil
}

func (c *SpotifyClient) CreatePlaylist(
	ctx context.Context,
	title string,
	description string,
) (*entities.CreatePlaylistOutput, error) {
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

	out := &entities.CreatePlaylistOutput{
		ID:  playlist.ID.String(),
		URL: playlist.ExternalURLs["spotify"],
	}

	return out, nil
}

func (c *SpotifyClient) AddTracksToPlaylist(
	ctx context.Context,
	input entities.AddTracksToPlaylistClientInput,
) error {
	c.Logger.Debug("Adding tracks to playlist...", map[string]interface{}{
		"playlist_id": input.PlaylistID,
		"song_ids":    input.Tracks,
	})

	if _, err := c.AuthenticatedClient.AddTracksToPlaylist(ctx, input.GetPlaylistID(), input.GetTrackIDs()...); err != nil {
		return err
	}

	return nil
}
