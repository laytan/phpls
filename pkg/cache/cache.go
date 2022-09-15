package cache

import (
	"fmt"
	"log"

	"github.com/DmitriyVTitov/size"
	"github.com/davecgh/go-spew/spew"
	lru "github.com/hashicorp/golang-lru"
	"github.com/laytan/elephp/pkg/datasize"
)

// Wrapper of the arc cache, adding generics and a size in bytes (approx).
type Cache[K comparable, V any] struct {
	capacity int
	c        *lru.Cache
}

// Creates a new cache with some sane defaults.
func New[K comparable, V any](capacity int) *Cache[K, V] {
	log.Printf("Creating arc cache of capacity %d\n", capacity)

	c, err := lru.New(capacity)
	if err != nil {
		panic(err)
	}

	return &Cache[K, V]{
		capacity: capacity,
		c:        c,
	}
}

// Puts the given key&value into the cache.
func (c *Cache[K, V]) Put(key K, value V) {
	c.c.Add(key, value)
}

// Gets the given key from the cache.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	value, ok := c.c.Get(key)

	return value.(V), ok
}

func (c *Cache[K, V]) Delete(key K) {
	c.c.Remove(key)
}

// Convenience function that gets the entry for the given key, if it is not
// in the cache, it calls valueCreator and returns what it returns, adding
// the entry into the cache for next calls.
func (c *Cache[K, V]) Cached(key K, valueCreator func() V) V {
	value, ok := c.c.Get(key)
	if ok {
		return value.(V)
	}

	created := valueCreator()

	c.c.Add(key, created)
	return created
}

func (c *Cache[K, V]) Length() int {
	return c.c.Len()
}

// Returns some stats about the cache, this is relatively expensive because we
// have to calculate the memory usage of the struct (not fast).
func (c *Cache[K, V]) String() string {
	s := datasize.Size(size.Of(c) * datasize.BitsInByte)
	sizeStats := fmt.Sprintf("Cache size %d/%d(%s)\n", c.c.Len(), c.capacity, s.String())

	return sizeStats
}

// Returns the same output as String() but with the keys that are cached at the
// end.
func (c *Cache[K, V]) StringWithKeys() string {
	sizeStats := c.String()
	keys := spew.Sprint(c.c.Keys())

	return sizeStats + "\nKeys: " + keys
}
