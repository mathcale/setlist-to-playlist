package setlistfm

import (
	"fmt"
	"net/url"
	"strings"
)

type ExtractIDFromURLUseCaseInterface interface {
	Execute(url string) (*string, error)
}

type ExtractIDFromURLUseCase struct{}

func NewExtractIDFromURLUseCase() ExtractIDFromURLUseCaseInterface {
	return &ExtractIDFromURLUseCase{}
}

func (uc *ExtractIDFromURLUseCase) Execute(u string) (*string, error) {
	if u == "" {
		return nil, fmt.Errorf("URL is empty")
	}

	if !strings.Contains(u, "https://www.setlist.fm/setlist") {
		return nil, fmt.Errorf("URL is not a valid setlist.fm set")
	}

	parsedURL, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	path := parsedURL.EscapedPath()
	splittedPath := strings.Split(path, "/")
	noDashes := strings.Split(splittedPath[len(splittedPath)-1], "-")
	id := strings.Replace(noDashes[len(noDashes)-1], ".html", "", -1)

	return &id, nil
}
