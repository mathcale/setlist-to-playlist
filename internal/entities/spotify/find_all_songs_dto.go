package spotify

type Song struct {
	ID    string
	Title string
	Album string
}

type FindAllSongsOutput struct {
	Artist string
	Songs  []Song
}
