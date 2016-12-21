package cache

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type keyProviderTest int

func (kpt keyProviderTest) Key() string {
	return fmt.Sprintf("%d", kpt)
}

func TestKey(t *testing.T) {
	s := "foobar"
	b := []byte("foobar")
	assert.Equal(t, "foobar", Key(s))
	assert.Equal(t, "foobar", Key(&s))
	assert.Equal(t, "123", Key(keyProviderTest(123)))
	assert.Equal(t, "foobar", Key(b))
	assert.Equal(t, "foobar", Key(&b))
	assert.Equal(t, "1.23", Key(float64(1.23)))
}
