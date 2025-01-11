package httpc

import (
	"net/http"
)

type Middleware interface {
	RoundTripper(next http.RoundTripper) http.RoundTripper
}

type MiddlewareFunc func(next http.RoundTripper) http.RoundTripper

func (f MiddlewareFunc) RoundTripper(next http.RoundTripper) http.RoundTripper {
	return f(next)
}

type RoundTripperFunc func(req *http.Request) (*http.Response, error)

func (f RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

var _ http.RoundTripper = (*Transport)(nil)

type Transport struct {
	rt          http.RoundTripper
	middlewares []Middleware
}

// NewTransport creates a new instance of CustomTransport.
func NewTransport(opts ...Option) *Transport {
	rt := clonedTransport(http.DefaultTransport)
	t := &Transport{
		rt: rt,
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// Option is a functional option for configuring the CustomTransport.
type Option func(*Transport)

// WithBaseTransport sets the base RoundTripper for the CustomTransport.
func WithBaseTransport(rt http.RoundTripper) Option {
	return func(t *Transport) {
		t.rt = rt
	}
}

// clonedTransport returns the given RoundTripper as a cloned *http.Transport.
// It returns nil if the RoundTripper can't be cloned.
func clonedTransport(rt http.RoundTripper) *http.Transport {
	t, ok := rt.(interface {
		Clone() *http.Transport
	})
	if !ok {
		return nil
	}
	return t.Clone()
}

// RoundTrip executes a single HTTP transaction, as defined by http.RoundTripper.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(t.middlewares) == 0 {
		return t.rt.RoundTrip(req)
	}

	var next http.RoundTripper = t.rt
	for i := len(t.middlewares) - 1; i >= 0; i-- {
		next = t.middlewares[i].RoundTripper(next)
	}
	return next.RoundTrip(req)
}

// Use adds the provided middleware functions to the CustomTransport's middleware chain.
func (t *Transport) Use(middleware ...Middleware) *Transport {
	t.middlewares = append(t.middlewares, middleware...)
	return t
}
