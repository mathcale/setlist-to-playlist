package music

import (
	"context"
	"errors"
	"strings"
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

	s.Run("picks the best candidate of each song, keeping setlist order", func() {
		s.ProviderMock.On("SearchCandidates", mock.Anything, "a", "artist").Return([]music.Candidate{
			{TrackID: "a1", Title: "A", Artist: "artist", Album: "Album A", AlbumType: music.AlbumTypeAlbum},
			{TrackID: "a2", Title: "A (Live)", Artist: "artist", Album: "Live", AlbumType: music.AlbumTypeSingle},
		}, nil).Once()
		s.ProviderMock.On("SearchCandidates", mock.Anything, "b", "artist").Return([]music.Candidate{
			{TrackID: "b1", Title: "B", Artist: "artist", Album: "Album B", AlbumType: music.AlbumTypeUnknown},
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

	pick := func(artist string, cs ...music.Candidate) *music.FetchSongsOutput {
		s.ProviderMock.ExpectedCalls = nil
		s.ProviderMock.On("SearchCandidates", mock.Anything, "t", artist).Return(cs, nil).Once()
		out, err := s.FetchSongs.Execute(ctx, []string{"t"}, artist)
		s.NoError(err)
		return out
	}
	cand := func(id, title, album string, t music.AlbumType) music.Candidate {
		return music.Candidate{TrackID: id, Title: title, Artist: "Artist", Album: album, AlbumType: t}
	}
	ids := func(o *music.FetchSongsOutput) []string {
		r := []string{}
		for _, x := range o.Songs {
			r = append(r, x.ID)
		}
		return r
	}

	s.Run("drops candidates by other artists, matching case-insensitively and trimmed", func() {
		out := pick("  ARTIST ",
			music.Candidate{TrackID: "other", Title: "t", Artist: "Someone Else", Album: "x", AlbumType: music.AlbumTypeAlbum},
			cand("ok", "t", "x", music.AlbumTypeSingle),
		)
		s.Equal([]string{"ok"}, ids(out))
	})

	s.Run("skips the song when only other artists remain", func() {
		out := pick("Artist", music.Candidate{TrackID: "o", Title: "t", Artist: "Other", AlbumType: music.AlbumTypeAlbum})
		s.Empty(out.Songs)
		s.Equal([]string{"t"}, out.Skipped)
	})

	s.Run("ranks album/EP > compilation > single > unknown > live", func() {
		live := cand("live", "t - Live", "x", music.AlbumTypeAlbum)
		all := []music.Candidate{
			live,
			cand("unk", "t", "x", music.AlbumTypeUnknown),
			cand("single", "t", "x", music.AlbumTypeSingle),
			cand("comp", "t", "x", music.AlbumTypeCompilation),
			cand("ep", "t", "x", music.AlbumTypeEP),
		}
		for _, want := range []string{"ep", "comp", "single", "unk", "live"} {
			s.Equal([]string{want}, ids(pick("Artist", all...)))
			rest := []music.Candidate{}
			for _, c := range all {
				if c.TrackID != want {
					rest = append(rest, c)
				}
			}
			all = rest
		}
	})

	s.Run("first wins within a tier", func() {
		out := pick("Artist",
			cand("single", "t", "x", music.AlbumTypeSingle),
			cand("ep", "t", "x", music.AlbumTypeEP),
			cand("album", "t", "x", music.AlbumTypeAlbum),
		)
		s.Equal([]string{"ep"}, ids(out))
	})

	s.Run("detects live markers in title or album, case-insensitively", func() {
		for _, m := range []string{"(Live", "- Live", "Live at", "Live from", "Unplugged", "Session"} {
			for _, inTitle := range []bool{true, false} {
				l := cand("live", "t", "x", music.AlbumTypeAlbum)
				if inTitle {
					l.Title = "t " + strings.ToUpper(m)
				} else {
					l.Album = "x " + strings.ToLower(m)
				}
				out := pick("Artist", l, cand("studio", "t", "x", music.AlbumTypeSingle))
				s.Equal([]string{"studio"}, ids(out), "%s inTitle=%v", m, inTitle)
			}
		}
	})

	s.Run("uses the best live candidate when all are live", func() {
		a := cand("liveSingle", "t (Live)", "x", music.AlbumTypeSingle)
		b := cand("liveAlbum", "t", "Live at Wembley", music.AlbumTypeAlbum)
		out := pick("Artist", a, b)
		s.Equal([]string{"liveSingle"}, ids(out))
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
