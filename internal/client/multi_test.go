package client

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/matryer/is"
)

type clifail int

func (clifail) ExpireTime(_ context.Context, domain string) (time.Time, error) {
	return time.Time{}, errors.New("foo")
}

type clisuccess time.Time

func (c clisuccess) ExpireTime(_ context.Context, domain string) (time.Time, error) {
	return time.Time(c), nil
}

func TestMulti(t *testing.T) {
	ctx := context.Background()
	t.Run("first client succeed", func(t *testing.T) {
		is := is.New(t)
		expected := time.Now()
		expire, err := NewMultiClient(clisuccess(expected), clifail(0)).ExpireTime(ctx, "a")
		is.NoErr(err)              // expected no error
		is.Equal(expected, expire) // expeted the same result
	})
	t.Run("last client succeed", func(t *testing.T) {
		is := is.New(t)
		expected := time.Now()
		expire, err := NewMultiClient(clifail(0), clifail(0), clisuccess(expected)).ExpireTime(ctx, "a")
		is.NoErr(err)              // expected no error
		is.Equal(expected, expire) // expeted the same result
	})
	t.Run("no client succeed", func(t *testing.T) {
		is := is.New(t)
		expire, err := NewMultiClient(clifail(0), clifail(0), clifail(0)).ExpireTime(ctx, "a")
		is.True(err != nil)           // expected an error
		is.Equal(err.Error(), "foo")  // expected the correct error msg
		is.Equal(expire, time.Time{}) // expected a zeroed result
	})
}
