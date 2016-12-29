[![Build Status](https://semaphoreci.com/api/v1/brad/gocache/branches/master/shields_badge.svg)](https://semaphoreci.com/brad/gocache)
[![codecov](https://codecov.io/gh/bradberger/gocache/branch/master/graph/badge.svg)](https://codecov.io/gh/bradberger/gocache)

Cache is a package which implements a unified interface for various Golang based cache.

Current drivers include:

- In-memory LRU cache  [github.com/bradberger/gocache/drivers/lru](https://godoc.org/github.com/bradberger/gocache/drivers/lru)
- [memcached](https://godoc.org/github.com/bradfitz/gomemcache/memcache)
- Redis via [redigo](https://godoc.org/github.com/garyburd/redigo/redis)
- [github.com/patricknm/go-cache](https://godoc.org/github.com/patricknm/go-cache)
- [App Engine memcache](https://godoc.org/google.golang.org/appengine/memcache)
- [github.com/armon/go-radix](https://godoc.org/github.com/armon/go-radix)

More drivers are most welcome! Just make sure they meet at least the `"cache".Cache`
interface and are unit tested.
