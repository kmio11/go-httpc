package cachemw

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/kmio11/go-httpc"
	"golang.org/x/sync/singleflight"
)

type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte) error
}

type CacheWithTTL interface {
	Cache
	SetWithTTL(ctx context.Context, key string, value []byte, ttl time.Duration) error
}

// KeyGenerator is an interface that defines a method for generating a cache key
// based on an HTTP request.
// If a Cache implementation also implements this interface, the CacheMiddleware
// will use the Key method to generate cache keys.
type KeyGenerator interface {
	Key(req *http.Request) (string, error)
}

var _ httpc.Middleware = (*CacheMiddleware)(nil)

type CacheMiddleware struct {
	group        singleflight.Group
	cache        Cache
	cacheableReq func(req *http.Request) bool
	ttl          time.Duration
}

type Option func(*CacheMiddleware)

func cacheableReq(req *http.Request) bool {
	return true
}

func GenerateRequestHash(req *http.Request) (string, error) {
	method := req.Method
	url := req.URL.String()
	var body []byte
	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return "", err
		}
		req.Body = io.NopCloser(bytes.NewReader(body))
	}

	hash := fnv.New64a()
	hash.Write([]byte(method))
	hash.Write([]byte(url))
	hash.Write(body)
	return fmt.Sprintf("%x", hash.Sum64()), nil
}

func New(cache Cache, opts ...Option) *CacheMiddleware {
	c := &CacheMiddleware{
		group:        singleflight.Group{},
		cache:        cache,
		cacheableReq: cacheableReq,
		ttl:          0,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithCacheableReq is an option function that sets a custom function to determine
// if a request is cacheable. The provided function `f` takes an *http.Request
// and returns a boolean indicating whether the request should be cached.
func WithCacheableReq(f func(req *http.Request) bool) Option {
	return func(c *CacheMiddleware) {
		c.cacheableReq = f
	}
}

// WithTTL sets the time-to-live (TTL) duration for the cache middleware.
// The TTL determines how long cached items should be retained before they expire.
func WithTTL(ttl time.Duration) Option {
	return func(c *CacheMiddleware) {
		c.ttl = ttl
	}
}

func (m *CacheMiddleware) RoundTripper(next http.RoundTripper) http.RoundTripper {
	return httpc.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		ctx := req.Context()

		if !m.cacheableReq(req) {
			return next.RoundTrip(req)
		}

		key, err := m.key(req)
		if err != nil {
			return next.RoundTrip(req)
		}

		if cached, err := m.cache.Get(ctx, key); err == nil {
			if resp, err := m.bytesToResponse(cached, req); err == nil {
				return resp, nil
			}
		}

		maybeResp, err, _ := m.group.Do(key, func() (any, error) {
			resp, err := next.RoundTrip(req)
			if err != nil {
				return resp, err
			}

			if dumpedResp, err := httputil.DumpResponse(resp, true); err == nil {
				if c, ok := m.cache.(CacheWithTTL); ok && m.ttl > 0 {
					_ = c.SetWithTTL(ctx, key, dumpedResp, m.ttl)
				} else {
					_ = m.cache.Set(ctx, key, dumpedResp)
				}
			}

			return resp, nil
		})

		return maybeResp.(*http.Response), err
	})
}

func (m *CacheMiddleware) bytesToResponse(data []byte, req *http.Request) (*http.Response, error) {
	return http.ReadResponse(bufio.NewReader(bytes.NewReader(data)), req)
}

func (m *CacheMiddleware) key(req *http.Request) (string, error) {
	if g, ok := m.cache.(KeyGenerator); ok {
		return g.Key(req)
	}
	return GenerateRequestHash(req)
}
