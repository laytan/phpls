package lfudacache

import "github.com/laytan/elephp/pkg/datasize"

// An entry in the cache.
type Entry[K comparable, V any] struct {
	Key   K
	Value V

	size                     datasize.Size
	lastUsedAge              uint
	timesUsed                uint
	scoreAgeMultiplier       uint
	scoreFrequencyMultiplier uint

	Prev *Entry[K, V]
	Next *Entry[K, V]
}

// Calculates the score for the item.
func (e *Entry[K, V]) Score() uint {
	return (e.lastUsedAge * e.scoreAgeMultiplier) * (e.timesUsed * e.scoreFrequencyMultiplier)
}
