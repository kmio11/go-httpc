package textcache

import (
	"context"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"

	"github.com/kmio11/go-httpc/cachemw"
)

var _ interface {
	cachemw.Cache
	cachemw.KeyGenerator
} = (*TextCache)(nil)

type TextCache struct {
	baseDir string
	dumpReq bool
}

type Option func(*TextCache)

func New(baseDir string, opts ...Option) *TextCache {
	c := &TextCache{
		baseDir: baseDir,
		dumpReq: false,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func WithDumpRequest() Option {
	return func(c *TextCache) {
		c.dumpReq = true
	}
}

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (c *TextCache) dumpReqFilepath(key string) string {
	return filepath.Join(c.baseDir, key, "request.txt")
}

func (c *TextCache) Key(req *http.Request) (string, error) {
	key, err := cachemw.GenerateRequestHash(req)
	if err != nil {
		return "", err
	}
	if !c.dumpReq {
		return key, nil
	}

	// dump request
	dump, err := httputil.DumpRequest(req, true)
	if err != nil {
		return key, err
	}
	if err := writeFile(c.dumpReqFilepath(key), dump); err != nil {
		return key, err
	}
	return key, nil
}

func (c *TextCache) cacheFilepath(key string) string {
	return filepath.Join(c.baseDir, key, "response.txt")
}

func (c *TextCache) Get(ctx context.Context, key string) ([]byte, error) {
	return os.ReadFile(c.cacheFilepath(key))
}

func (c *TextCache) Set(ctx context.Context, key string, value []byte) error {
	return writeFile(c.cacheFilepath(key), value)
}
