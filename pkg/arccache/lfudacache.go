package arccache

import (
	"fmt"
	"log"
	"time"

	"github.com/DmitriyVTitov/size"
	"github.com/davecgh/go-spew/spew"
	lru "github.com/hashicorp/golang-lru"
	"github.com/laytan/elephp/pkg/datasize"
)

const (
	// Average IR tree size to base the capacity of the cache on.
	// actually calculating the size of the cache (bytes) every time
	// is too slow, so we aproximate.
	avgIRTreeSize = datasize.MegaByte * 3
)

// Wrapper of the arc cache, adding generics and a size in bytes (approx).
type Cache[K comparable, V any] struct {
	targetSize datasize.Size
	c          *lru.ARCCache
}

// Creates a new cache with some sane defaults.
func New[K comparable, V any](targetSize datasize.Size) *Cache[K, V] {
	capacity := int(targetSize / avgIRTreeSize)
	log.Printf("Creating arc cache of capacity %d\n", capacity)

	c, err := lru.NewARC(capacity)
	if err != nil {
		panic(err)
	}

	cache := &Cache[K, V]{
		targetSize: targetSize,
		c:          c,
	}

	t := time.NewTicker(time.Minute)
	go func() {
		for {
			<-t.C

			log.Println(cache.String())
		}
	}()

	return cache
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

// Returns a string representing the state of the cache, useful in logging.
func (c *Cache[K, V]) String() string {
	capacity := int(c.targetSize / avgIRTreeSize)
	s := datasize.Size(size.Of(c) * datasize.BitsInByte)
	sizeStats := fmt.Sprintf("Cache size %d/%d(%s)\n", c.c.Len(), capacity, s.String())
	keys := spew.Sprint(c.c.Keys())

	return sizeStats + "\nKeys: " + keys
}
