package ch08

import (
	"net/http"
	"testing"
	"time"
)

func TestHeadTime(t *testing.T) {
	resp, err := http.Head("https://www.time.gov/")
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close() // Always close the body

	now := time.Now().Round(time.Second)

	date := resp.Header.Get("Date")
	if date == "" {
		t.Fatal("no Date header received from time.gov")
	}

	time, err := time.Parse(time.RFC1123, date)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("time.gov: %s (skew %s)", time, now.Sub(time))
}
