package spotify

import (
	"context"
	"net/http"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"

	oauth2util "github.com/mathcale/setlist-to-playlist/internal/pkg/oauth2"
)

type SpotifyClientInterface interface {
	NewAPIClient(ctx context.Context, tok *oauth2.Token) *spotify.Client
	GetToken(ctx context.Context, r *http.Request, state string, genCodes oauth2util.GenerateOutput) (*oauth2.Token, error)
	GetAuthURL(state string, genCodes oauth2util.GenerateOutput) string
}

type SpotifyClient struct {
	Auth              *spotifyauth.Authenticator
	PKCECodeGenerator oauth2util.PKCECodeGeneratorInterface
	ClientID          string
	ClientSecret      string
}

func NewSpotifyClient(
	gen oauth2util.PKCECodeGeneratorInterface,
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
		PKCECodeGenerator: gen,
	}
}

func (c *SpotifyClient) NewAPIClient(ctx context.Context, tok *oauth2.Token) *spotify.Client {
	return spotify.New(c.Auth.Client(ctx, tok))
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

func (c *SpotifyClient) GetAuthURL(state string, genCodes oauth2util.GenerateOutput) string {
	return c.Auth.AuthURL(
		state,
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("code_challenge", genCodes.CodeChallenge),
	)
}
