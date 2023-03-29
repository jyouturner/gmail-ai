package integration

import (
	"net/http"
	"strconv"
	"time"

	"github.com/PuerkitoBio/rehttp"
)

// RoundTripFunc is an adapter to allow the use of ordinary functions as HTTP
// round trips.
type RoundTripFunc func(*http.Request) (*http.Response, error)

// RoundTrip executes a single HTTP transaction, returning
// a Response for the provided Request.
func (rf RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return rf(req)
}

// RateLimitTransport wraps base transport with rate limiting functionality.
//
// When a 429 status code is returned by the remote server, the
// "X-RateLimit-Reset" header is used to determine how long the transport will
// wait until re-issuing the failed request.
func RateLimitTransport(base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return rehttp.NewTransport(base, retry, delay)
}

func retry(attempt rehttp.Attempt) bool {
	if attempt.Response == nil {
		return false
	}
	return attempt.Response.StatusCode == http.StatusTooManyRequests
}

func delay(attempt rehttp.Attempt) time.Duration {
	resetAt := attempt.Response.Header.Get("X-RateLimit-Reset")
	resetAtUnix, err := strconv.ParseInt(resetAt, 10, 64)
	if err != nil {
		resetAtUnix = time.Now().Add(5 * time.Second).Unix()
	}
	return time.Duration(resetAtUnix-time.Now().Unix()) * time.Second
}

// Option is the type used to configure a client.
type Option func(*http.Client)

// WithRateLimit configures the client to enable rate limiting.
func WithRateLimit() Option {
	return func(c *http.Client) {
		c.Transport = RateLimitTransport(c.Transport)
	}
}

// Wrap the base client.
func Wrap(base *http.Client, options ...Option) *http.Client {
	if base == nil {
		base = http.DefaultClient
	}
	client := &http.Client{
		Timeout:   base.Timeout,
		Transport: base.Transport,
	}
	for _, option := range options {
		option(client)
	}
	return client
}
