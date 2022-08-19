package lfudacache

import (
	"fmt"
	"strings"
	"sync"

	"github.com/DmitriyVTitov/size"
	"github.com/laytan/elephp/pkg/datasize"
	log "github.com/sirupsen/logrus"
)

const (
	defaultMinItemsInCache          = 4
	defaultScoreAgeMultiplier       = 2
	defaultScoreFrequencyMultiplier = 1
)

// Implementation of a LFUDA cache.
// With the added benefit of generics and having the size based of actual bytes.
// Which 3rd party packages did not really have.
type Cache[K comparable, V any] struct {
	// Max size in bytes for the cache.
	targetSize datasize.Size
	// Makes sure that entries could theoretically fit MinItemsInCache into the cache
	// if they don't, they won't get added at all.
	minItemsInCache uint

	// Multipliers to control the importance of age versus that of frequency.
	scoreAgeMultiplier       uint
	scoreFrequencyMultiplier uint

	mu sync.Mutex

	currentSize datasize.Size
	currentAge  uint

	lookup map[K]*Entry[K, V]
	head   *Entry[K, V]
	tail   *Entry[K, V]
}

// Creates a new cache with some sane defaults.
func New[K comparable, V any](targetSize datasize.Size) *Cache[K, V] {
	return &Cache[K, V]{
		targetSize:               targetSize,
		lookup:                   make(map[K]*Entry[K, V]),
		minItemsInCache:          defaultMinItemsInCache,
		scoreAgeMultiplier:       defaultScoreAgeMultiplier,
		scoreFrequencyMultiplier: defaultScoreFrequencyMultiplier,
	}
}

// Puts the given key&value into the cache.
func (c *Cache[K, V]) Put(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if map has it,
	// If it does, remove it but keep the timesused for the new entry.
	timesUsed := uint(1)
	if entry, ok := c.lookup[key]; ok {
		timesUsed = entry.timesUsed

		c.remove(entry)
	}

	bSize := datasize.Size(size.Of(value))
	if uint(bSize) >= uint(c.targetSize)/c.minItemsInCache {
		log.Warnf(
			"Size of entry: %s for key %v is too large, not putting it in cache\n",
			bSize.String(),
			key,
		)
		return
	}

	// Keep removing until we have space for the new entry.
	for {
		newSize := c.currentSize + bSize
		if newSize <= c.targetSize {
			break
		}

		log.Infof("Removed entry with key %v out of cache\n", c.tail.Key)
		c.remove(c.tail)
	}

	c.insert(&Entry[K, V]{
		size:                     bSize,
		lastUsedAge:              c.currentAge,
		timesUsed:                timesUsed,
		Key:                      key,
		Value:                    value,
		scoreAgeMultiplier:       c.scoreAgeMultiplier,
		scoreFrequencyMultiplier: c.scoreFrequencyMultiplier,
	})
	c.currentAge++

	log.Infof("New size of cache: %s\n", c.currentSize.String())
}

// Gets the given key from the cache.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, ok := c.lookup[key]; ok {
		entry.lastUsedAge = c.currentAge
		entry.timesUsed++
		c.currentAge++

		c.remove(entry)
		c.insert(entry)

		return entry.Value, true
	}

	var defaultV V
	return defaultV, false
}

// Convenience function that gets the entry for the given key, if it is not
// in the cache, it calls valueCreator and returns what it returns, adding
// the entry into the cache for next calls.
func (c *Cache[K, V]) Cached(key K, valueCreator func() V) V {
	if v, ok := c.Get(key); ok {
		return v
	}

	v := valueCreator()
	c.Put(key, v)
	return v
}

func (c *Cache[K, V]) Length() int {
	var length int
	for curr := c.head; curr != nil; curr = curr.Next {
		length++
	}

	return length
}

// Returns a string representing the state of the cache, useful in logging.
func (c *Cache[K, V]) String() string {
	results := []string{}
	for curr := c.head; curr != nil; curr = curr.Next {
		results = append(results, fmt.Sprintf("%d: %v", curr.Score(), curr.Key))
	}

	return fmt.Sprintf(
		"Age: %d, Count: %d, Size of items: %s, Max size %s | %s",
		c.currentAge,
		c.Length(),
		c.currentSize.String(),
		c.targetSize.String(),
		strings.Join(results, ", "),
	)
}

func (c *Cache[K, V]) remove(entry *Entry[K, V]) {
	defer func() {
		c.currentSize -= entry.size
		delete(c.lookup, entry.Key)
	}()

	if entry == c.head {
		c.head = c.head.Next
		if c.head != nil {
			c.head.Prev = nil
		}

		return
	}

	if entry == c.tail {
		c.tail = c.tail.Prev
		if c.tail != nil {
			c.tail.Next = nil
		}

		return
	}

	if entry.Next != nil {
		entry.Next.Prev = entry.Prev
	}

	if entry.Prev != nil {
		entry.Prev.Next = entry.Next
	}
}

func (c *Cache[K, V]) insert(entry *Entry[K, V]) {
	defer func() {
		c.lookup[entry.Key] = entry
		c.currentSize += entry.size
	}()

	if c.head == nil {
		c.head = entry
		c.tail = entry
		return
	}

	// Go from head to tail until an entry is found with a lower score,
	// Put it in front of that one.
	score := entry.Score()
	var insertAfter *Entry[K, V]
	for curr := c.head; curr != nil; curr = curr.Next {
		if score >= curr.Score() {
			insertAfter = curr.Prev
			break
		}
	}

	if insertAfter == nil {
		t := c.head
		c.head = entry
		entry.Next = t
		t.Prev = entry

		return
	}

	if insertAfter.Next != nil {
		entry.Next = insertAfter.Next
	}

	insertAfter.Next = entry
	entry.Prev = insertAfter
}
