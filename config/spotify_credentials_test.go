package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/mathcale/setlist-to-playlist/internal/tests/mocks"
)

func loadWith(t *testing.T, toml string) ConfigPaths {
	t.Helper()
	viper.Reset()
	t.Cleanup(viper.Reset)

	dir := t.TempDir()
	assert.NoError(t, os.WriteFile(dir+"/config.toml", []byte(toml), 0600))
	paths := ConfigPaths{AppConfigDir: dir, AppConfigFile: dir + "/config.toml"}

	// Loading app config must never prompt (setlistfm key is already set).
	_, err := Load(paths, &mocks.PrompterMock{})
	assert.NoError(t, err)

	return paths
}

func TestSpotifyCredentialStore_FirstUsePromptsAndSaves(t *testing.T) {
	paths := loadWith(t, "[setlistfm]\napi_key='k'\n")

	p := &mocks.PrompterMock{}
	p.On("Text", "What's your Spotify client ID?").Return("the-id", nil).Once()
	p.On("Secret", "What's your Spotify client secret?").Return("the-secret", nil).Once()

	store := NewSpotifyCredentialStore(p)
	assert.NoError(t, store.Ensure())

	id, secret := store.Credentials()
	assert.Equal(t, "the-id", id)
	assert.Equal(t, "the-secret", secret)
	p.AssertExpectations(t)

	raw, err := os.ReadFile(paths.AppConfigFile)
	assert.NoError(t, err)
	assert.Contains(t, string(raw), "the-id")
	assert.Contains(t, string(raw), "the-secret")
}

func TestSpotifyCredentialStore_DoesNotPromptOnceSaved(t *testing.T) {
	loadWith(t, "[setlistfm]\napi_key='k'\n[spotify]\nclient_id='i'\nclient_secret='s'\n")

	p := &mocks.PrompterMock{}
	store := NewSpotifyCredentialStore(p)

	assert.NoError(t, store.Ensure())
	assert.NoError(t, store.Ensure())

	p.AssertNumberOfCalls(t, "Text", 0)
	p.AssertNumberOfCalls(t, "Secret", 0)
}

func TestSpotifyCredentialStore_PromptsOnlyForMissingSecret(t *testing.T) {
	loadWith(t, "[setlistfm]\napi_key='k'\n[spotify]\nclient_id='i'\n")

	p := &mocks.PrompterMock{}
	p.On("Secret", "What's your Spotify client secret?").Return("s", nil).Once()

	assert.NoError(t, NewSpotifyCredentialStore(p).Ensure())
	p.AssertNumberOfCalls(t, "Text", 0)
	p.AssertExpectations(t)
}

func TestSpotifyCredentialStore_RejectsEmptyInput(t *testing.T) {
	paths := loadWith(t, "[setlistfm]\napi_key='k'\n")

	p := &mocks.PrompterMock{}
	p.On("Text", "What's your Spotify client ID?").Return("  ", nil).Once()

	err := NewSpotifyCredentialStore(p).Ensure()

	assert.ErrorContains(t, err, "Spotify client ID cannot be empty")
	p.AssertNumberOfCalls(t, "Secret", 0)

	raw, readErr := os.ReadFile(paths.AppConfigFile)
	assert.NoError(t, readErr)
	assert.NotContains(t, string(raw), "client_id")
}
