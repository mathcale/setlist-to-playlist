package persistence

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mathcale/setlist-to-playlist/internal/entities/applemusic"
)

type fakeStrategy struct {
	data     []byte
	readErr  error
	writeErr error
}

func (f *fakeStrategy) Read() ([]byte, error) { return f.data, f.readErr }
func (f *fakeStrategy) Write(d []byte) error {
	f.data = d
	return f.writeErr
}

func TestAppleMusicAuthPersistence(t *testing.T) {
	t.Run("round-trips auth data", func(t *testing.T) {
		p := NewAppleMusicAuthPersistence(&fakeStrategy{}, nil)
		in := applemusic.AuthData{BearerToken: "b", MediaUserToken: "m"}

		assert.NoError(t, p.Write(in))

		out, err := p.Read()
		assert.NoError(t, err)
		assert.Equal(t, in, *out)
	})

	t.Run("reads the empty placeholder file as empty data", func(t *testing.T) {
		out, err := NewAppleMusicAuthPersistence(&fakeStrategy{data: []byte("{}")}, nil).Read()
		assert.NoError(t, err)
		assert.Error(t, out.Validate())
	})

	t.Run("returns read, parse and write errors", func(t *testing.T) {
		_, err := NewAppleMusicAuthPersistence(&fakeStrategy{readErr: errors.New("r")}, nil).Read()
		assert.EqualError(t, err, "r")

		_, err = NewAppleMusicAuthPersistence(&fakeStrategy{data: []byte("nope")}, nil).Read()
		assert.Error(t, err)

		err = NewAppleMusicAuthPersistence(&fakeStrategy{writeErr: errors.New("w")}, nil).Write(applemusic.AuthData{})
		assert.EqualError(t, err, "w")
	})
}
