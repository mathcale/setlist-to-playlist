package spotify

import "github.com/zmb3/spotify/v2"

type AddTracksToPlaylistInput struct {
	PlaylistID string
	Tracks     []Song
}

type AddTracksToPlaylistClientInput struct {
	PlaylistID string
	Tracks     []Song
}

func (cin AddTracksToPlaylistClientInput) GetPlaylistID() spotify.ID {
	return spotify.ID(cin.PlaylistID)
}

func (cin AddTracksToPlaylistClientInput) GetTrackIDs() []spotify.ID {
	songIDs := make([]spotify.ID, len(cin.Tracks))

	for i, s := range cin.Tracks {
		songIDs[i] = spotify.ID(s.ID)
	}

	return songIDs
}
