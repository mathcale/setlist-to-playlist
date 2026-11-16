package music

import (
	"context"
	"errors"
	"fmt"

	"github.com/mathcale/setlist-to-playlist/internal/entities/music"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/logger"
)

// Recovery adds credential recovery to provider calls. When a call fails with
// music.ErrCredentialsExpired it tells the user, re-authenticates and retries
// that single call once. If the retry (or the re-authentication) fails with
// expired credentials again, it returns an error carrying Help. Other errors
// are never retried.
//
// "Exactly once" is scoped per provider call: each call gets at most one
// re-authentication plus one retry. Because every call is wrapped on its own,
// a multi-step operation never repeats finished steps (e.g. only AddTracks is
// retried after CreatePlaylist succeeded, and only the failing search is
// retried while fetching songs).
type Recovery struct {
	Auth   music.Authenticator
	Logger logger.LoggerInterface
	Name   string // provider display name
	Help   string // how to get fresh credentials
}

func NewRecovery(a music.Authenticator, l logger.LoggerInterface, id music.ProviderID) *Recovery {
	return &Recovery{Auth: a, Logger: l, Name: id.DisplayName(), Help: id.CredentialsHelp()}
}

func (r *Recovery) failed() error {
	return fmt.Errorf("%w: %s still rejected the credentials; %s", music.ErrCredentialsExpired, r.Name, r.Help)
}

func (r *Recovery) reauth(ctx context.Context) error {
	r.Logger.Warn(fmt.Sprintf("Your %s credentials have expired, please provide new ones", r.Name), nil)

	if err := r.Auth.Reauthenticate(ctx); err != nil {
		if errors.Is(err, music.ErrCredentialsExpired) {
			return r.failed()
		}

		return err
	}

	return nil
}

// Authenticate runs the initial authentication. An expired-credentials result
// triggers one re-authentication (which validates the new credentials itself).
func (r *Recovery) Authenticate(ctx context.Context) error {
	err := r.Auth.Authenticate(ctx)
	if !errors.Is(err, music.ErrCredentialsExpired) {
		return err
	}

	return r.reauth(ctx)
}

// Do runs op, recovering once from expired credentials.
func (r *Recovery) Do(ctx context.Context, op func() error) error {
	err := op()
	if !errors.Is(err, music.ErrCredentialsExpired) {
		return err
	}

	if err := r.reauth(ctx); err != nil {
		return err
	}

	err = op()
	if errors.Is(err, music.ErrCredentialsExpired) {
		return r.failed()
	}

	return err
}

// RecoveringProvider wraps a MusicProvider so every call goes through Recovery.
type RecoveringProvider struct {
	Provider music.MusicProvider
	Recovery *Recovery
}

var _ music.MusicProvider = (*RecoveringProvider)(nil)

func NewRecoveringProvider(p music.MusicProvider, r *Recovery) *RecoveringProvider {
	return &RecoveringProvider{Provider: p, Recovery: r}
}

func (p *RecoveringProvider) SearchCandidates(ctx context.Context, title string, artist string) ([]music.Candidate, error) {
	var out []music.Candidate

	err := p.Recovery.Do(ctx, func() (err error) {
		out, err = p.Provider.SearchCandidates(ctx, title, artist)
		return err
	})

	return out, err
}

func (p *RecoveringProvider) CreatePlaylist(ctx context.Context, title string, description string) (*music.Playlist, error) {
	var out *music.Playlist

	err := p.Recovery.Do(ctx, func() (err error) {
		out, err = p.Provider.CreatePlaylist(ctx, title, description)
		return err
	})

	return out, err
}

func (p *RecoveringProvider) AddTracks(ctx context.Context, playlistID string, trackIDs []string) error {
	return p.Recovery.Do(ctx, func() error {
		return p.Provider.AddTracks(ctx, playlistID, trackIDs)
	})
}
