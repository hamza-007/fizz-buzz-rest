package stats

import "sync"

// TODO: per-process and lost on restart; move behind a shared store (Redis...).
type Counter[K comparable] struct {
	mu      sync.Mutex
	counts  map[K]uint64
	maxKeys int

	top     K
	topHits uint64
}

func NewCounter[K comparable](maxKeys int) *Counter[K] {
	return &Counter[K]{
		counts:  make(map[K]uint64),
		maxKeys: maxKeys,
	}
}

// TODO: one mutex per served request serialises the API; shard the map by key hash.
func (c *Counter[K]) Add(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	hits, tracked := c.counts[key]
	if !tracked && c.maxKeys > 0 && len(c.counts) >= c.maxKeys {
		return
	}

	hits++
	c.counts[key] = hits

	if hits > c.topHits {
		c.top, c.topHits = key, hits
	}
}

func (c *Counter[K]) Top() (key K, hits uint64, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.topHits == 0 {
		var zero K
		return zero, 0, false
	}
	return c.top, c.topHits, true
}
