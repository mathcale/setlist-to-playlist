package spotify

import "strings"

type Song struct {
	ID    string
	Title string
	Album string
}

type FindAllSongsOutput struct {
	Artist string
	Songs  []Song
}

func (faso *FindAllSongsOutput) AddSong(song Song) {
	faso.Songs = append(faso.Songs, song)
}

func (faso *FindAllSongsOutput) SortSongs(original []string) {
	sorted := make([]Song, len(original))

	for i, s := range original {
		for _, song := range faso.Songs {
			if strings.Contains(song.Title, s) {
				sorted[i] = song
			}
		}
	}

	faso.Songs = sorted
}
