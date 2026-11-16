package commands

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/setlist-to-playlist/internal/entities/music"
	"github.com/mathcale/setlist-to-playlist/internal/entities/setlistfm"
	"github.com/mathcale/setlist-to-playlist/internal/tests/mocks"
)

type RootCmdTestSuite struct {
	suite.Suite
	LoggerMock         *mocks.LoggerMock
	RootCmdGatewayMock *mocks.RootCmdGatewayMock

	Cmd RootCmdInterface
}

func (s *RootCmdTestSuite) SetupTest() {
	s.LoggerMock = new(mocks.LoggerMock)
	s.RootCmdGatewayMock = new(mocks.RootCmdGatewayMock)

	s.Cmd = NewRootCmd(
		s.LoggerMock,
		s.RootCmdGatewayMock,
	)
}

func (s *RootCmdTestSuite) cleanMocks() {
	s.LoggerMock.ExpectedCalls = nil
	s.LoggerMock.Calls = nil
	s.RootCmdGatewayMock.ExpectedCalls = nil
	s.RootCmdGatewayMock.Calls = nil
}

func TestRootCmd(t *testing.T) {
	suite.Run(t, new(RootCmdTestSuite))
}

func (s *RootCmdTestSuite) TestBuild() {
	s.Run("Should build a new command", func() {
		cmd := s.Cmd.Build()

		s.NotNil(cmd)
	})

	s.Run("should have the correct short description", func() {
		cmd := s.Cmd.Build()

		s.Equal("Creates a playlist based on a Setlist.fm entry", cmd.Short)
	})

	s.Run("Should have the correct run function", func() {
		cmd := s.Cmd.Build()

		s.NotNil(cmd.RunE)
	})

	s.Run("Should have the correct flags", func() {
		cmd := s.Cmd.Build()

		flags := cmd.Flags()

		s.NotNil(flags.Lookup("url"))
		s.NotNil(flags.Lookup("provider"))
		s.Equal("p", flags.Lookup("provider").Shorthand)
		s.Equal("spotify", flags.Lookup("provider").DefValue)
	})
}

func (s *RootCmdTestSuite) TestRun() {
	url := "https://www.setlist.fm/setlist/blink182/2024/autodromo-de-interlagos-sao-paulo-brazil-53aa1325.html"

	s.Run("Should execute the run function", func() {
		defer s.cleanMocks()

		set := &setlistfm.Set{
			ID: "any-set-id",
			Sets: setlistfm.Sets{
				Set: []setlistfm.Songs{
					{
						Song: []setlistfm.Song{
							{Name: "any-song-1"},
							{Name: "any-song-2"},
							{Name: "any-song-3"},
						},
					},
					{
						Song: []setlistfm.Song{
							{Name: "any-song-4"},
						},
						Encore: 1,
					},
				},
			},
		}

		songs := &music.FetchSongsOutput{
			Artist: "any-artist",
			Songs: []music.Song{
				{ID: "any-song-id-1", Title: "any-song-1", Album: "any-album-1"},
				{ID: "any-song-id-2", Title: "any-song-2", Album: "any-album-1"},
				{ID: "any-song-id-3", Title: "any-song-3", Album: "any-album-1"},
			},
		}

		playlistURL := "https://open.spotify.com/playlist/any-playlist-id"

		s.LoggerMock.On("Info", mock.Anything, mock.Anything).Return()
		s.RootCmdGatewayMock.On("CheckProvider", music.ProviderSpotify).Return(nil)
		s.RootCmdGatewayMock.On("GetTracksFromSetlist", mock.Anything).Return(set, nil)
		s.RootCmdGatewayMock.On("Authenticate", mock.Anything, music.ProviderSpotify).Return(nil)
		s.RootCmdGatewayMock.
			On("FetchSongs", mock.Anything, music.ProviderSpotify, set.Songs(), set.ArtistName()).
			Return(songs, nil)
		s.RootCmdGatewayMock.
			On("CreatePlaylist", mock.Anything, music.ProviderSpotify, set.Title(), songs.Songs).
			Return(&playlistURL, nil)

		cmd := s.Cmd.Build()
		err := cmd.RunE(cmd, []string{
			"--url", url,
		})

		expectedMsg := fmt.Sprintf("Playlist created successfully, check it out: %s", playlistURL)

		s.NoError(err)
		s.Equal(s.LoggerMock.Calls[len(s.LoggerMock.Calls)-1].Arguments[0].(string), expectedMsg)
	})

	s.Run("Should return an error when failing to extract Setlist.fm ID from URL", func() {
		defer s.cleanMocks()

		s.LoggerMock.On("Info", mock.Anything, mock.Anything).Return()
		s.RootCmdGatewayMock.On("CheckProvider", music.ProviderSpotify).Return(nil)
		s.LoggerMock.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()
		s.RootCmdGatewayMock.On("GetTracksFromSetlist", mock.Anything).Return(nil, errors.New("any-validation-error"))

		cmd := s.Cmd.Build()
		err := cmd.RunE(cmd, []string{
			"--url", url,
		})

		s.Error(err)
		s.ErrorContains(err, "any-validation-error")
	})

	s.Run("Should return an error when failing to get Setlist.fm data", func() {
		defer s.cleanMocks()

		s.LoggerMock.On("Info", mock.Anything, mock.Anything).Return()
		s.RootCmdGatewayMock.On("CheckProvider", music.ProviderSpotify).Return(nil)
		s.LoggerMock.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()
		s.RootCmdGatewayMock.
			On("GetTracksFromSetlist", mock.Anything).
			Return(nil, errors.New("any-error"))

		cmd := s.Cmd.Build()
		err := cmd.RunE(cmd, []string{
			"--url", url,
		})

		s.Error(err)
		s.ErrorContains(err, "any-error")
	})

	s.Run("Should return an error when failing to authenticate on Spotify", func() {
		defer s.cleanMocks()

		s.LoggerMock.On("Info", mock.Anything, mock.Anything).Return()
		s.RootCmdGatewayMock.On("CheckProvider", music.ProviderSpotify).Return(nil)
		s.LoggerMock.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()
		s.RootCmdGatewayMock.On("GetTracksFromSetlist", mock.Anything).Return(&setlistfm.Set{}, nil)
		s.RootCmdGatewayMock.
			On("Authenticate", mock.Anything, music.ProviderSpotify).
			Return(errors.New("any-error"))

		cmd := s.Cmd.Build()
		err := cmd.RunE(cmd, []string{
			"--url", url,
		})

		s.Error(err)
		s.ErrorContains(err, "any-error")
	})

	s.Run("Should return an error when failing to fetch songs from Spotify", func() {
		defer s.cleanMocks()

		set := &setlistfm.Set{
			ID: "any-set-id",
			Sets: setlistfm.Sets{
				Set: []setlistfm.Songs{
					{
						Song: []setlistfm.Song{
							{Name: "any-song-1"},
							{Name: "any-song-2"},
							{Name: "any-song-3"},
						},
					},
					{
						Song: []setlistfm.Song{
							{Name: "any-song-4"},
						},
						Encore: 1,
					},
				},
			},
		}

		s.LoggerMock.On("Info", mock.Anything, mock.Anything).Return()
		s.RootCmdGatewayMock.On("CheckProvider", music.ProviderSpotify).Return(nil)
		s.LoggerMock.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()
		s.RootCmdGatewayMock.On("GetTracksFromSetlist", mock.Anything).Return(set, nil)
		s.RootCmdGatewayMock.On("Authenticate", mock.Anything, music.ProviderSpotify).Return(nil)
		s.RootCmdGatewayMock.
			On("FetchSongs", mock.Anything, music.ProviderSpotify, set.Songs(), set.ArtistName()).
			Return(nil, errors.New("any-error"))

		cmd := s.Cmd.Build()
		err := cmd.RunE(cmd, []string{
			"--url", url,
		})

		s.Error(err)
		s.ErrorContains(err, "any-error")
	})

	s.Run("Should return an error when failing to create a playlist on Spotify", func() {
		defer s.cleanMocks()

		set := &setlistfm.Set{
			ID: "any-set-id",
			Sets: setlistfm.Sets{
				Set: []setlistfm.Songs{
					{
						Song: []setlistfm.Song{
							{Name: "any-song-1"},
							{Name: "any-song-2"},
							{Name: "any-song-3"},
						},
					},
					{
						Song: []setlistfm.Song{
							{Name: "any-song-4"},
						},
						Encore: 1,
					},
				},
			},
		}

		songs := &music.FetchSongsOutput{
			Artist: "any-artist",
			Songs: []music.Song{
				{ID: "any-song-id-1", Title: "any-song-1", Album: "any-album-1"},
				{ID: "any-song-id-2", Title: "any-song-2", Album: "any-album-1"},
				{ID: "any-song-id-3", Title: "any-song-3", Album: "any-album-1"},
			},
		}

		s.LoggerMock.On("Info", mock.Anything, mock.Anything).Return()
		s.RootCmdGatewayMock.On("CheckProvider", music.ProviderSpotify).Return(nil)
		s.LoggerMock.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()
		s.RootCmdGatewayMock.On("GetTracksFromSetlist", mock.Anything).Return(set, nil)
		s.RootCmdGatewayMock.On("Authenticate", mock.Anything, music.ProviderSpotify).Return(nil)
		s.RootCmdGatewayMock.
			On("FetchSongs", mock.Anything, music.ProviderSpotify, set.Songs(), set.ArtistName()).
			Return(songs, nil)
		s.RootCmdGatewayMock.
			On("CreatePlaylist", mock.Anything, music.ProviderSpotify, set.Title(), songs.Songs).
			Return(nil, errors.New("any-error"))

		cmd := s.Cmd.Build()
		err := cmd.RunE(cmd, []string{
			"--url", url,
		})

		s.Error(err)
		s.ErrorContains(err, "any-error")
	})

	s.Run("Should accept --provider and the -p shorthand", func() {
		for _, flag := range []string{"--provider", "-p"} {
			cmd := s.Cmd.Build()
			s.NoError(cmd.ParseFlags([]string{flag, "apple-music", "--url", url}))

			got, _ := cmd.Flags().GetString("provider")
			s.Equal("apple-music", got)
		}
	})

	s.Run("Should default to spotify when the flag is absent", func() {
		defer s.cleanMocks()

		s.LoggerMock.On("Info", mock.Anything, mock.Anything).Return()
		s.RootCmdGatewayMock.On("CheckProvider", music.ProviderSpotify).Return(nil)
		s.RootCmdGatewayMock.On("GetTracksFromSetlist", mock.Anything).Return(nil, errors.New("stop"))
		s.LoggerMock.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

		cmd := s.Cmd.Build()
		_ = cmd.ParseFlags([]string{"--url", url})
		_ = cmd.RunE(cmd, nil)

		s.RootCmdGatewayMock.AssertCalled(s.T(), "CheckProvider", music.ProviderSpotify)
	})

	s.Run("Should return an error listing valid providers for an invalid value", func() {
		defer s.cleanMocks()

		cmd := s.Cmd.Build()
		_ = cmd.ParseFlags([]string{"--url", url, "-p", "tidal"})
		err := cmd.RunE(cmd, nil)

		s.ErrorContains(err, "spotify, apple-music")
		s.RootCmdGatewayMock.AssertNotCalled(s.T(), "CheckProvider", mock.Anything)
	})

	s.Run("Should exit cleanly without authenticating when the provider is unavailable", func() {
		defer s.cleanMocks()

		s.LoggerMock.On("Info", mock.Anything, mock.Anything).Return()
		s.RootCmdGatewayMock.
			On("CheckProvider", music.ProviderAppleMusic).
			Return(fmt.Errorf("Apple Music: %w", music.ErrProviderUnavailable))

		cmd := s.Cmd.Build()
		_ = cmd.ParseFlags([]string{"--url", url, "-p", "apple-music"})
		err := cmd.RunE(cmd, nil)

		s.NoError(err)
		s.Contains(s.LoggerMock.Calls[0].Arguments[0].(string), "Apple Music support is not available yet")
		s.RootCmdGatewayMock.AssertNotCalled(s.T(), "Authenticate", mock.Anything, mock.Anything)
	})

	s.Run("Should name the provider in progress logs", func() {
		defer s.cleanMocks()

		s.LoggerMock.On("Info", mock.Anything, mock.Anything).Return()
		s.RootCmdGatewayMock.On("CheckProvider", music.ProviderSpotify).Return(nil)
		s.RootCmdGatewayMock.On("GetTracksFromSetlist", mock.Anything).Return(&setlistfm.Set{}, nil)
		s.RootCmdGatewayMock.On("Authenticate", mock.Anything, music.ProviderSpotify).Return(nil)
		s.RootCmdGatewayMock.On("FetchSongs", mock.Anything, music.ProviderSpotify, mock.Anything, mock.Anything).
			Return(&music.FetchSongsOutput{}, nil)
		url := "u"
		s.RootCmdGatewayMock.On("CreatePlaylist", mock.Anything, music.ProviderSpotify, mock.Anything, mock.Anything).
			Return(&url, nil)

		cmd := s.Cmd.Build()
		_ = cmd.ParseFlags([]string{"--url", "x"})
		s.NoError(cmd.RunE(cmd, nil))

		var msgs []string
		for _, c := range s.LoggerMock.Calls {
			msgs = append(msgs, c.Arguments[0].(string))
		}

		s.Contains(msgs, "Authenticating on Spotify...")
		s.Contains(msgs, "Fetching songs on Spotify...")
		s.Contains(msgs, "Creating playlist on Spotify...")
	})
}
