package setlistfm

import (
	"fmt"

	"github.com/mathcale/setlist-to-playlist/internal/entities/setlistfm"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/httpclient"
)

type SetlistFMClientInterface interface {
	GetSetlistByID(setlistID string) (*setlistfm.Set, error)
}

type SetlistFMClient struct {
	HttpClient httpclient.HttpClientInterface
	APIKey     string
}

var (
	GetSetlistByIDPath = "/1.0/setlist/%s"
)

func NewSetlistFMClient(httpClient httpclient.HttpClientInterface, apiKey string) SetlistFMClientInterface {
	return &SetlistFMClient{
		HttpClient: httpClient,
		APIKey:     apiKey,
	}
}

func (c *SetlistFMClient) GetSetlistByID(id string) (*setlistfm.Set, error) {
	var setlist setlistfm.Set

	headers := map[string]interface{}{
		"x-api-key": c.APIKey,
	}

	err := c.HttpClient.Get(fmt.Sprintf(GetSetlistByIDPath, id), headers, &setlist)
	if err != nil {
		return nil, err
	}

	return &setlist, nil
}
