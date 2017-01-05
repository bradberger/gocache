package kv

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/bradberger/gocache/cache"
	"github.com/bradberger/gokv/codec"
	dv "github.com/bradberger/gokv/drivers/diskv"
	"github.com/peterbourgon/diskv"

	"github.com/stretchr/testify/assert"
)

func getTestOptions() diskv.Options {
	tmpDir, err := ioutil.TempDir("", fmt.Sprintf("diskv"))
	if err != nil {
		panic(err)
	}
	return diskv.Options{
		BasePath:     tmpDir,
		Transform:    func(s string) []string { return []string{} },
		CacheSizeMax: 1024 * 1024,
	}
}

func TestGetSetDel(t *testing.T) {

	opts := getTestOptions()
	db := dv.New(opts)
	defer func() {
		os.RemoveAll(opts.BasePath)
	}()

	c := New(db)

	var s string
	assert.Equal(t, cache.ErrCacheMiss, c.Get("foo", &s))
	assert.NoError(t, c.Set("foo", "bar", 100*time.Millisecond))
	assert.True(t, c.Exists("foo"))
	assert.NoError(t, c.Get("foo", &s))
	assert.Equal(t, "bar", s)
	time.Sleep(150 * time.Millisecond)
	assert.False(t, c.Exists("foo"))

	origCodec := dv.Codec
	defer func() {
		dv.Codec = origCodec
	}()
	dv.Codec = codec.ErrTestCodec
	assert.Error(t, c.Set("foo", "bar", 0))

}
