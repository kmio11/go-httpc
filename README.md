# httpc

A Go library for creating and customizing HTTP clients. It provides:

- A pluggable `Transport` that lets you attach multiple middlewares to process requests and responses.
- A `Middleware` interface for adding reusable handlers like logging or caching.
- Built-in caching middlewares, with options for various backends (e.g. [`redis`](cachemw/rediscache/redis.go), [`badger`](cachemw/badgercache/badger.go), [`text file`](cachemw/textcache/text.go)).
- An easy way to add request/response hooks for debugging or logging.

See [example_test.go](example_test.go) for usage examples.