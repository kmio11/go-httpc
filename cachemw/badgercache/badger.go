package badgercache

import (
	"context"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/kmio11/go-httpc/cachemw"
)

var _ cachemw.CacheWithTTL = (*BadgerCache)(nil)

type BadgerCache struct {
	db *badger.DB
}

type Option func(*BadgerCache)

func New(db *badger.DB, opts ...Option) *BadgerCache {
	c := &BadgerCache{
		db: db,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *BadgerCache) Get(ctx context.Context, key string) ([]byte, error) {
	var resp []byte
	err := c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		resp, err = item.ValueCopy(nil)
		return err
	})
	return resp, err
}

func (c *BadgerCache) setEntry(_ context.Context, entry *badger.Entry) error {
	return c.db.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(entry)
	})
}

func (c *BadgerCache) Set(ctx context.Context, key string, value []byte) error {
	return c.setEntry(ctx,
		badger.NewEntry([]byte(key), value),
	)
}

func (c *BadgerCache) SetWithTTL(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.setEntry(ctx,
		badger.NewEntry([]byte(key), value).WithTTL(ttl),
	)
}
