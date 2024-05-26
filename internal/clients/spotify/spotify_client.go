package spotify

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"

	entities "github.com/mathcale/setlist-to-playlist/internal/entities/spotify"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/logger"
	oauth2util "github.com/mathcale/setlist-to-playlist/internal/pkg/oauth2"
)

type SpotifyClientInterface interface {
	GetToken(ctx context.Context, r *http.Request, state string, genCodes oauth2util.GenerateOutput) (*oauth2.Token, error)
	GetAuthURL(state string, genCodes oauth2util.GenerateOutput) string
	NewAPIClient(ctx context.Context, tok *oauth2.Token) *spotify.Client
	SetAuthenticatedClient(ch chan *spotify.Client)
	CurrentUser(ctx context.Context) (*spotify.PrivateUser, error)
	CurrentSession() (*oauth2.Token, error)
	FindAllSongsByName(ctx context.Context, name []string, artist string) (*entities.FindAllSongsOutput, error)
	CreatePlaylist(ctx context.Context, title string, description string) (*entities.CreatePlaylistOutput, error)
	AddTracksToPlaylist(ctx context.Context, input entities.AddTracksToPlaylistClientInput) error
}

type SpotifyClient struct {
	Auth                *spotifyauth.Authenticator
	AuthenticatedClient *spotify.Client
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
		AuthenticatedClient: nil,
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

func (c *SpotifyClient) SetAuthenticatedClient(ch chan *spotify.Client) {
	c.AuthenticatedClient = <-ch
}

func (c *SpotifyClient) CurrentUser(ctx context.Context) (*spotify.PrivateUser, error) {
	return c.AuthenticatedClient.CurrentUser(ctx)
}

func (c *SpotifyClient) CurrentSession() (*oauth2.Token, error) {
	return c.AuthenticatedClient.Token()
}

func (c *SpotifyClient) FindAllSongsByName(
	ctx context.Context,
	name []string,
	artist string,
) (*entities.FindAllSongsOutput, error) {
	// TODO: use goroutines to search for each song in parallel
	result := &entities.FindAllSongsOutput{
		Artist: artist,
	}

	for _, n := range name {
		q := fmt.Sprintf(`"%s"%%20artist:%s`, strings.ToLower(n), strings.ToLower(artist))

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

			return nil, err
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

			result.Songs = append(result.Songs, song)
		}
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
