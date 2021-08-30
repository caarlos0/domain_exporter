package refresher

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeOk struct {
}

func (fakeOk) ExpireTime(ctx context.Context, domain string) (time.Time, error) {
	return time.Time{}, nil
}

type fakeFail struct {
}

func (fakeFail) ExpireTime(ctx context.Context, domain string) (time.Time, error) {
	return time.Time{}, errors.New("foo")
}

func Test_refresher_Refresh(t *testing.T) {
	tests := []struct {
		name      string
		refresher refresher
	}{
		{
			name:      "refresh is ok",
			refresher: New(time.Second, fakeOk{}, "foo.com"),
		},
		{
			name:      "refresh is failed",
			refresher: New(time.Second, fakeFail{}, "foo.com"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.refresher.Refresh(context.Background())
		})
	}
}
