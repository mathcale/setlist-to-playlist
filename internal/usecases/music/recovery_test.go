package music

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mathcale/setlist-to-playlist/internal/entities/music"
	"github.com/mathcale/setlist-to-playlist/internal/tests/mocks"
)

func expired() error { return errors.Join(music.ErrCredentialsExpired, errors.New("http 401")) }

type recFixture struct {
	provider *mocks.MusicProviderMock
	auth     *mocks.AuthenticatorMock
	logger   *mocks.LoggerMock
	p        *RecoveringProvider
	rec      *Recovery
}

func newRecFixture() *recFixture {
	f := &recFixture{
		provider: new(mocks.MusicProviderMock),
		auth:     new(mocks.AuthenticatorMock),
		logger:   new(mocks.LoggerMock),
	}
	f.logger.On("Warn", mock.Anything, mock.Anything).Return()
	f.logger.On("Debug", mock.Anything, mock.Anything).Return()
	f.rec = NewRecovery(f.auth, f.logger, music.ProviderAppleMusic)
	f.p = NewRecoveringProvider(f.provider, f.rec)

	return f
}

func TestRecoveringProvider(t *testing.T) {
	ctx := context.Background()

	t.Run("retries once after re-authenticating and succeeds", func(t *testing.T) {
		f := newRecFixture()
		f.provider.On("SearchCandidates", mock.Anything, "t", "a").Return(nil, expired()).Once()
		f.provider.On("SearchCandidates", mock.Anything, "t", "a").Return([]music.Candidate{{TrackID: "1"}}, nil).Once()
		f.auth.On("Reauthenticate", mock.Anything).Return(nil).Once()

		got, err := f.p.SearchCandidates(ctx, "t", "a")

		assert.NoError(t, err)
		assert.Len(t, got, 1)
		f.auth.AssertNumberOfCalls(t, "Reauthenticate", 1)
		f.provider.AssertNumberOfCalls(t, "SearchCandidates", 2)
		f.logger.AssertNumberOfCalls(t, "Warn", 1)
	})

	t.Run("second expiry returns an error with devtools instructions", func(t *testing.T) {
		f := newRecFixture()
		f.provider.On("CreatePlaylist", mock.Anything, "t", "d").Return(nil, expired())
		f.auth.On("Reauthenticate", mock.Anything).Return(nil).Once()

		_, err := f.p.CreatePlaylist(ctx, "t", "d")

		assert.ErrorIs(t, err, music.ErrCredentialsExpired)
		assert.ErrorContains(t, err, "devtools")
		assert.ErrorContains(t, err, "media-user-token")
		assert.ErrorContains(t, err, "music.apple.com")
		f.provider.AssertNumberOfCalls(t, "CreatePlaylist", 2)
		f.auth.AssertNumberOfCalls(t, "Reauthenticate", 1)
	})

	t.Run("re-authentication failing with expiry returns the instructions", func(t *testing.T) {
		f := newRecFixture()
		f.provider.On("AddTracks", mock.Anything, "p", []string{"1"}).Return(expired())
		f.auth.On("Reauthenticate", mock.Anything).Return(expired())

		err := f.p.AddTracks(ctx, "p", []string{"1"})

		assert.ErrorContains(t, err, "devtools")
		f.provider.AssertNumberOfCalls(t, "AddTracks", 1)
	})

	t.Run("other re-authentication errors are returned as is", func(t *testing.T) {
		f := newRecFixture()
		boom := errors.New("prompt aborted")
		f.provider.On("AddTracks", mock.Anything, mock.Anything, mock.Anything).Return(expired())
		f.auth.On("Reauthenticate", mock.Anything).Return(boom)

		assert.Equal(t, boom, f.p.AddTracks(ctx, "p", nil))
	})

	t.Run("non-auth errors are not retried", func(t *testing.T) {
		f := newRecFixture()
		boom := errors.New("http 500")
		f.provider.On("SearchCandidates", mock.Anything, mock.Anything, mock.Anything).Return(nil, boom)

		_, err := f.p.SearchCandidates(ctx, "t", "a")

		assert.Equal(t, boom, err)
		f.provider.AssertNumberOfCalls(t, "SearchCandidates", 1)
		f.auth.AssertNotCalled(t, "Reauthenticate", mock.Anything)
	})

	t.Run("fetch songs retries only the failing search", func(t *testing.T) {
		f := newRecFixture()
		f.provider.On("SearchCandidates", mock.Anything, "one", "a").Return([]music.Candidate{}, nil).Once()
		f.provider.On("SearchCandidates", mock.Anything, "two", "a").Return(nil, expired()).Once()
		f.provider.On("SearchCandidates", mock.Anything, "two", "a").Return([]music.Candidate{}, nil).Once()
		f.auth.On("Reauthenticate", mock.Anything).Return(nil).Once()

		_, err := NewFetchSongsUseCase(f.p, f.logger).Execute(ctx, []string{"one", "two"}, "a")

		assert.NoError(t, err)
		f.provider.AssertNumberOfCalls(t, "SearchCandidates", 3)
	})

	t.Run("add tracks retry does not create a second playlist", func(t *testing.T) {
		f := newRecFixture()
		f.provider.On("CreatePlaylist", mock.Anything, mock.Anything, mock.Anything).Return(&music.Playlist{ID: "p"}, nil).Once()
		f.provider.On("AddTracks", mock.Anything, "p", mock.Anything).Return(expired()).Once()
		f.provider.On("AddTracks", mock.Anything, "p", mock.Anything).Return(nil).Once()
		f.auth.On("Reauthenticate", mock.Anything).Return(nil).Once()

		_, err := NewCreatePlaylistUseCase(f.p, f.logger).Execute(ctx, music.CreatePlaylistInput{Title: "t"})
		assert.NoError(t, err)
		assert.NoError(t, NewAddTracksUseCase(f.p, f.logger).Execute(ctx, music.AddTracksInput{PlaylistID: "p", Tracks: []music.Song{{ID: "1"}}}))

		f.provider.AssertNumberOfCalls(t, "CreatePlaylist", 1)
		f.provider.AssertNumberOfCalls(t, "AddTracks", 2)
	})
}

func TestRecoveryAuthenticate(t *testing.T) {
	ctx := context.Background()

	t.Run("expired validation triggers re-authentication", func(t *testing.T) {
		f := newRecFixture()
		f.auth.On("Authenticate", mock.Anything).Return(expired())
		f.auth.On("Reauthenticate", mock.Anything).Return(nil)

		assert.NoError(t, f.rec.Authenticate(ctx))
		f.auth.AssertNumberOfCalls(t, "Authenticate", 1)
	})

	t.Run("second expiry returns the instructions", func(t *testing.T) {
		f := newRecFixture()
		f.auth.On("Authenticate", mock.Anything).Return(expired())
		f.auth.On("Reauthenticate", mock.Anything).Return(expired())

		assert.ErrorContains(t, f.rec.Authenticate(ctx), "devtools")
	})

	t.Run("other errors are not retried", func(t *testing.T) {
		f := newRecFixture()
		f.auth.On("Authenticate", mock.Anything).Return(errors.New("boom"))

		assert.EqualError(t, f.rec.Authenticate(ctx), "boom")
		f.auth.AssertNotCalled(t, "Reauthenticate", mock.Anything)
	})
}
