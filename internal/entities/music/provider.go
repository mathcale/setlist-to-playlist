package music

import (
	"errors"
	"fmt"
	"strings"
)

// ProviderID identifies a streaming service, as typed on the --provider flag.
type ProviderID string

const (
	ProviderSpotify    ProviderID = "spotify"
	ProviderAppleMusic ProviderID = "apple-music"
)

// DefaultProvider is used when --provider is absent.
const DefaultProvider = ProviderSpotify

// ProviderIDs lists every provider the CLI knows about, registered or not.
func ProviderIDs() []ProviderID {
	return []ProviderID{ProviderSpotify, ProviderAppleMusic}
}

// DisplayName is the human-facing name used in logs and messages.
func (id ProviderID) DisplayName() string {
	switch id {
	case ProviderSpotify:
		return "Spotify"
	case ProviderAppleMusic:
		return "Apple Music"
	default:
		return string(id)
	}
}

// ParseProviderID validates a raw flag value, returning an error that lists
// the valid values when it is unknown.
func ParseProviderID(raw string) (ProviderID, error) {
	valid := make([]string, 0, len(ProviderIDs()))

	for _, id := range ProviderIDs() {
		if string(id) == raw {
			return id, nil
		}

		valid = append(valid, string(id))
	}

	return "", fmt.Errorf("invalid provider %q, valid providers: %s", raw, strings.Join(valid, ", "))
}

// ErrProviderUnavailable is returned when a known provider has no
// implementation registered yet.
var ErrProviderUnavailable = errors.New("provider not available yet")

// ProviderEntry is the pair of ports needed to use one provider.
type ProviderEntry struct {
	Provider      MusicProvider
	Authenticator Authenticator
}

// ProviderRegistry maps provider IDs to their implementations.
type ProviderRegistry struct {
	entries map[ProviderID]ProviderEntry
}

func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{entries: map[ProviderID]ProviderEntry{}}
}

// Register adds a provider; it returns the registry so calls can be chained.
func (r *ProviderRegistry) Register(id ProviderID, p MusicProvider, a Authenticator) *ProviderRegistry {
	r.entries[id] = ProviderEntry{Provider: p, Authenticator: a}
	return r
}

// Lookup returns the entry for id, or ErrProviderUnavailable (wrapped with the
// display name) when none is registered.
func (r *ProviderRegistry) Lookup(id ProviderID) (ProviderEntry, error) {
	e, ok := r.entries[id]
	if !ok {
		return ProviderEntry{}, fmt.Errorf("%s: %w", id.DisplayName(), ErrProviderUnavailable)
	}

	return e, nil
}

// CredentialsHelp explains how to get fresh credentials for the provider. It
// is shown when re-authenticating did not fix an expired-credentials error.
func (id ProviderID) CredentialsHelp() string {
	switch id {
	case ProviderAppleMusic:
		return "get new tokens from https://music.apple.com (signed in): " +
			"the bearer token is the Authorization request header of any request to amp-api.music.apple.com " +
			"in devtools > Network (without the \"Bearer \" prefix), and the media-user-token is the cookie of the same " +
			"name in devtools > Application/Storage > Cookies > https://music.apple.com"
	default:
		return "re-run the command to authenticate again"
	}
}
