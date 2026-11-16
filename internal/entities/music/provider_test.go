package music

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseProviderID(t *testing.T) {
	for _, raw := range []string{"spotify", "apple-music"} {
		id, err := ParseProviderID(raw)
		assert.NoError(t, err)
		assert.Equal(t, ProviderID(raw), id)
	}

	_, err := ParseProviderID("tidal")
	assert.ErrorContains(t, err, "spotify, apple-music")
}

func TestDisplayName(t *testing.T) {
	assert.Equal(t, "Spotify", ProviderSpotify.DisplayName())
	assert.Equal(t, "Apple Music", ProviderAppleMusic.DisplayName())
}

func TestProviderRegistry(t *testing.T) {
	r := NewProviderRegistry()
	_, err := r.Lookup(ProviderSpotify)
	assert.True(t, errors.Is(err, ErrProviderUnavailable))

	r.Register(ProviderSpotify, nil, nil)
	_, err = r.Lookup(ProviderSpotify)
	assert.NoError(t, err)
}
