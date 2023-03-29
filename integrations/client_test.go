package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWrapRateLimit(t *testing.T) {

	start := time.Now()
	first := true
	// a mock server that returns a 429 on the first request and a 200 on the second
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if first {
			t.Log(start.Unix())
			w.Header().Set("X-RateLimit-Limit", "1")
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", fmt.Sprint(start.Add(time.Second).Unix()))
			w.WriteHeader(http.StatusTooManyRequests)
			first = !first
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	s := httptest.NewServer(h)
	defer s.Close()

	c := Wrap(s.Client(), WithRateLimit())
	r, err := c.Get(s.URL)
	if err != nil {
		t.Error(err)
	}

	if r.StatusCode != http.StatusOK {
		t.Errorf("Expected status code to be %d but got %d", http.StatusOK, r.StatusCode)
	}

	elapsed := time.Since(start)
	if elapsed < time.Second {
		t.Errorf("Time since start is sooner than expected. Expected >= 1s but got %s", elapsed)
	}
}
