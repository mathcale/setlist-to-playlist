package setlistfm

import (
	"errors"
	"net/url"
	"strings"
)

type GetSetlistByIDInput struct {
	URL string
}

func NewGetSetlistByIDInput(url string) GetSetlistByIDInput {
	return GetSetlistByIDInput{
		URL: url,
	}
}

func (in GetSetlistByIDInput) SetlistID() (*string, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}

	parsedURL, err := url.Parse(in.URL)
	if err != nil {
		return nil, err
	}

	path := parsedURL.EscapedPath()
	splittedPath := strings.Split(path, "/")
	noDashes := strings.Split(splittedPath[len(splittedPath)-1], "-")
	id := strings.Replace(noDashes[len(noDashes)-1], ".html", "", -1)

	return &id, nil
}

func (in GetSetlistByIDInput) Validate() error {
	if in.URL == "" {
		return errors.New("URL is empty")
	}

	if !strings.Contains(in.URL, "https://www.setlist.fm/setlist") {
		return errors.New("URL is not a valid setlist.fm set")
	}

	return nil
}
