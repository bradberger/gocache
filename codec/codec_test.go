package codec

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// gobErrStruct is a struct which should trigger a Gob encoding error
type gobErrStruct struct {
	sync.Mutex
}

func TestTestMarshalErr(t *testing.T) {
	b, err := testMarshalErr(nil)
	assert.Nil(t, b)
	assert.Error(t, err)
}

func TestTestUnmarshalErr(t *testing.T) {
	assert.Error(t, testUnmarshalErr(nil, nil))
}

func TestGobMarshal(t *testing.T) {
	v := 1.2
	b, err := gobMarshal(v)
	assert.NoError(t, err)
	assert.Equal(t, []byte{'\x0b', '\x08', '\x00', '\xf8', '\x33', '\x33', '\x33', '\x33', '\x33', '\x33', '\xf3', '\x3f'}, b)
}

func TestGobUnmarshal(t *testing.T) {
	var v float64
	b := []byte{'\x0b', '\x08', '\x00', '\xf8', '\x33', '\x33', '\x33', '\x33', '\x33', '\x33', '\xf3', '\x3f'}
	assert.NoError(t, gobUnmarshal(b, &v))
}

func TestGobMarshalErr(t *testing.T) {
	_, err := gobMarshal(&gobErrStruct{})
	assert.Error(t, err)
}
