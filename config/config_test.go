package config

import (
	"errors"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mathcale/setlist-to-playlist/internal/tests/mocks"
)

func TestLoad_FirstRunPromptsAndSavesValues(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	dir := t.TempDir()
	paths := ConfigPaths{AppConfigDir: dir, AppConfigFile: dir + "/config.toml"}

	p := &mocks.PrompterMock{}
	p.On("Text", "What's your Setlist.fm API key?").Return("key", nil)
	p.On("Text", "What's your Spotify client ID?").Return("id", nil)
	p.On("Text", "What's your Spotify client secret?").Return("secret", nil)

	c, err := Load(paths, p)

	assert.NoError(t, err)
	assert.Equal(t, "key", c.SetlistFM.APIKey)
	assert.Equal(t, "id", c.Spotify.ClientID)
	assert.Equal(t, "secret", c.Spotify.ClientSecret)
	p.AssertExpectations(t)

	raw, err := os.ReadFile(paths.AppConfigFile)
	assert.NoError(t, err)
	assert.Contains(t, string(raw), "secret")
}

func TestLoad_SkipsPromptsWhenConfigured(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	dir := t.TempDir()
	assert.NoError(t, os.WriteFile(dir+"/config.toml", []byte("[setlistfm]\napi_key='k'\n[spotify]\nclient_id='i'\nclient_secret='s'\n"), 0600))

	p := &mocks.PrompterMock{}
	_, err := Load(ConfigPaths{AppConfigDir: dir, AppConfigFile: dir + "/config.toml"}, p)

	assert.NoError(t, err)
	p.AssertNumberOfCalls(t, "Text", 0)
	p.AssertNumberOfCalls(t, "Secret", 0)
}

func TestLoad_ReturnsPromptError(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	dir := t.TempDir()
	p := &mocks.PrompterMock{}
	p.On("Text", mock.Anything).Return("", errors.New("aborted"))

	c, err := Load(ConfigPaths{AppConfigDir: dir, AppConfigFile: dir + "/config.toml"}, p)

	assert.Error(t, err)
	assert.Nil(t, c)
}
