package music

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/setlist-to-playlist/internal/entities/music"
	"github.com/mathcale/setlist-to-playlist/internal/tests/mocks"
)

type MusicUseCasesTestSuite struct {
	suite.Suite

	ProviderMock *mocks.MusicProviderMock
	LoggerMock   *mocks.LoggerMock

	FetchSongs     FetchSongsUseCaseInterface
	CreatePlaylist CreatePlaylistUseCaseInterface
	AddTracks      AddTracksUseCaseInterface
}

func (s *MusicUseCasesTestSuite) SetupTest() {
	s.ProviderMock = new(mocks.MusicProviderMock)
	s.LoggerMock = new(mocks.LoggerMock)
	s.LoggerMock.On("Debug", mock.Anything, mock.Anything).Return()

	s.FetchSongs = NewFetchSongsUseCase(s.ProviderMock, s.LoggerMock)
	s.CreatePlaylist = NewCreatePlaylistUseCase(s.ProviderMock, s.LoggerMock)
	s.AddTracks = NewAddTracksUseCase(s.ProviderMock, s.LoggerMock)
}

func TestMusicUseCases(t *testing.T) {
	suite.Run(t, new(MusicUseCasesTestSuite))
}

func (s *MusicUseCasesTestSuite) TestFetchSongs() {
	ctx := context.Background()

	s.Run("takes the first candidate of each song, keeping setlist order", func() {
		s.ProviderMock.On("SearchCandidates", mock.Anything, "a", "artist").Return([]music.Candidate{
			{TrackID: "a1", Title: "A", Album: "Album A", AlbumType: music.AlbumTypeAlbum},
			{TrackID: "a2", Title: "A (Live)", Album: "Live", AlbumType: music.AlbumTypeSingle},
		}, nil).Once()
		s.ProviderMock.On("SearchCandidates", mock.Anything, "b", "artist").Return([]music.Candidate{
			{TrackID: "b1", Title: "B", Album: "Album B", AlbumType: music.AlbumTypeUnknown},
		}, nil).Once()

		out, err := s.FetchSongs.Execute(ctx, []string{"a", "b"}, "artist")

		s.NoError(err)
		s.Equal("artist", out.Artist)
		s.Equal([]music.Song{
			{ID: "a1", Title: "A", Album: "Album A"},
			{ID: "b1", Title: "B", Album: "Album B"},
		}, out.Songs)
		s.Empty(out.Skipped)
	})

	s.Run("skips songs without candidates", func() {
		s.ProviderMock.On("SearchCandidates", mock.Anything, "none", "artist").Return([]music.Candidate{}, nil).Once()

		out, err := s.FetchSongs.Execute(ctx, []string{"none"}, "artist")

		s.NoError(err)
		s.Empty(out.Songs)
		s.Equal([]string{"none"}, out.Skipped)
	})

	s.Run("returns provider error", func() {
		s.ProviderMock.On("SearchCandidates", mock.Anything, "x", "artist").Return(nil, errors.New("any-error")).Once()

		out, err := s.FetchSongs.Execute(ctx, []string{"x"}, "artist")

		s.ErrorContains(err, "any-error")
		s.Nil(out)
	})
}

func (s *MusicUseCasesTestSuite) TestCreatePlaylist() {
	ctx := context.Background()

	s.Run("creates playlist with given description", func() {
		desc := "any-description"
		expected := &music.Playlist{ID: "id", URL: "url"}
		s.ProviderMock.On("CreatePlaylist", mock.Anything, "title", desc).Return(expected, nil).Once()

		out, err := s.CreatePlaylist.Execute(ctx, music.CreatePlaylistInput{Title: "title", Description: &desc})

		s.NoError(err)
		s.Equal(expected, out)
	})

	s.Run("falls back to the default description", func() {
		expected := &music.Playlist{ID: "id", URL: "url"}
		s.ProviderMock.On("CreatePlaylist", mock.Anything, "title", music.DefaultPlaylistDescription).Return(expected, nil).Once()

		out, err := s.CreatePlaylist.Execute(ctx, music.CreatePlaylistInput{Title: "title"})

		s.NoError(err)
		s.Equal(expected, out)
	})

	s.Run("returns provider error", func() {
		s.ProviderMock.On("CreatePlaylist", mock.Anything, "title", mock.Anything).Return(nil, errors.New("any-error")).Once()

		out, err := s.CreatePlaylist.Execute(ctx, music.CreatePlaylistInput{Title: "title"})

		s.ErrorContains(err, "any-error")
		s.Nil(out)
	})
}

func (s *MusicUseCasesTestSuite) TestAddTracks() {
	ctx := context.Background()
	input := music.AddTracksInput{
		PlaylistID: "pl",
		Tracks:     []music.Song{{ID: "t1"}, {ID: "t2"}, {ID: "t3"}},
	}

	s.Run("adds tracks in setlist order", func() {
		s.ProviderMock.On("AddTracks", mock.Anything, "pl", []string{"t1", "t2", "t3"}).Return(nil).Once()

		s.NoError(s.AddTracks.Execute(ctx, input))
		s.ProviderMock.AssertExpectations(s.T())
	})

	s.Run("returns provider error", func() {
		s.ProviderMock.On("AddTracks", mock.Anything, "pl", mock.Anything).Return(errors.New("any-error")).Once()

		s.ErrorContains(s.AddTracks.Execute(ctx, input), "any-error")
	})
}
