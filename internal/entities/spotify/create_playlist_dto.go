package spotify

var (
	DefaultPlaylistDescription = "Generated by 'Setlist to Playlist' script by @mathcale"
)

type CreatePlaylistInput struct {
	Title       string
	Description *string
}

type CreatePlaylistOutput struct {
	ID  string
	URL string
}

func (in CreatePlaylistInput) GetDescription() string {
	if in.Description == nil || *in.Description == "" {
		return DefaultPlaylistDescription
	}

	return *in.Description
}