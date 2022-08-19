package lfudacache

import (
	"fmt"
	"sort"
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

	entries map[K]*Entry[K, V]
	scores  []*Entry[K, V]
}

// Creates a new cache with some sane defaults.
func New[K comparable, V any](targetSize datasize.Size) *Cache[K, V] {
	return &Cache[K, V]{
		targetSize:               targetSize,
		entries:                  make(map[K]*Entry[K, V]),
		scores:                   make([]*Entry[K, V], 0),
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
	if entry, ok := c.entries[key]; ok {
		timesUsed = entry.timesUsed

		i := c.findIndex(entry)
		c.removeScore(i)

		c.currentSize -= entry.size
		delete(c.entries, key)
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

		removeKey := c.scores[0].Key
		removeEntry := c.entries[removeKey]

		c.scores = c.scores[1:len(c.scores)]
		c.currentSize -= removeEntry.size
		delete(c.entries, removeKey)
		log.Infof("Removed entry with key %v out of cache\n", removeKey)
	}

	entry := &Entry[K, V]{
		size:                     bSize,
		lastUsedAge:              c.currentAge,
		timesUsed:                timesUsed,
		Key:                      key,
		Value:                    value,
		scoreAgeMultiplier:       c.scoreAgeMultiplier,
		scoreFrequencyMultiplier: c.scoreFrequencyMultiplier,
	}

	i := c.findIndexFor(entry.Score())
	c.insertScore(i, entry)
	c.entries[entry.Key] = entry

	c.currentAge++
	c.currentSize += bSize
	log.Infof("New size of cache: %s\n", c.currentSize.String())
}

// Gets the given key from the cache.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, ok := c.entries[key]; ok {
		i := c.findIndex(entry)

		entry.lastUsedAge = c.currentAge
		entry.timesUsed++
		c.currentAge++

		// Re-insert it based on new score.
		c.removeScore(i)
		i = c.findIndexFor(entry.Score())
		c.insertScore(i, entry)

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
		log.Debugf("Cache HIT %v\n", key)
		return v
	}

	log.Debugf("Cache MISS %v\n", key)

	v := valueCreator()
	c.Put(key, v)
	return v
}

// Returns a string representing the state of the cache, useful in logging.
func (c *Cache[K, V]) String() string {
	results := make([]string, len(c.scores))
	for i, entry := range c.scores {
		results[i] = fmt.Sprintf("%d: %v", entry.Score(), entry.Key)
	}

	return fmt.Sprintf(
		"Age: %d, Count: %d, Size of items: %s, Max size %s | %s",
		c.currentAge,
		len(c.scores),
		c.currentSize.String(),
		c.targetSize.String(),
		strings.Join(results, ", "),
	)
}

// Binary searches for the index in scores of an entry.
// The given index would be the position for an entry with the given score.
func (c *Cache[K, V]) findIndexFor(score uint) int {
	// Find the first index that is at the same or higher score than given score.
	i := sort.Search(len(c.scores), func(i int) bool {
		return c.scores[i].Score() >= score
	})

	// Keep going untill we find an entry with a higher score, meaning we return the prev.
	for j := i; j < len(c.scores); j++ {
		if c.scores[j].Score() != score {
			return j
		}
	}

	// If there is no item higher than score, return the given i
	return i
}

// Binary searches for the index in scores of an entry.
// the entry at index will be the given entry.
func (c *Cache[K, V]) findIndex(entry *Entry[K, V]) int {
	score := entry.Score()
	// I is now the first index that is more than or equal to score.
	i := sort.Search(len(c.scores), func(i int) bool {
		return c.scores[i].Score() >= score
	})

	// Loop from that index, checking if the keys match, once we get out of
	// the score range, return, so that, we only check same score and same key.
	for j := i; j < len(c.scores); j++ {
		if c.scores[j].Score() != score {
			break
		}

		if c.scores[j].Key == entry.Key {
			return j
		}
	}

	return -1
}

func (c *Cache[K, V]) insertScore(index int, value *Entry[K, V]) {
	if len(c.scores) == index {
		c.scores = append(c.scores, value)
		return
	}

	c.scores = append(c.scores[:index+1], c.scores[index:]...)
	c.scores[index] = value
}

func (c *Cache[K, V]) removeScore(i int) {
	if len(c.scores)-1 == i {
		c.scores = c.scores[:i]
		return
	}

	c.scores = append(c.scores[:i], c.scores[i+1:]...)
}
