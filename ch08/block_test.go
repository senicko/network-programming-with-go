package ch08

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func blockIndifinitely(w http.ResponseWriter, r *http.Request) {
	select {}
}

func TestBlockIndifinitely(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(blockIndifinitely))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	defer func() {
		_ = resp.Body.Close()
	}()

	if err != nil {
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
			return
		}
		return
	}
}
