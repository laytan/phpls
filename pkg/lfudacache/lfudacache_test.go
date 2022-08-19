package lfudacache

import (
	"testing"

	"github.com/laytan/elephp/pkg/datasize"
	"github.com/matryer/is"
)

func TestPut(t *testing.T) {
	is := is.New(t)

	// Create a cache that can hold 3 bytes which is 3 int8's.
	c := New[string, int8](3)
	c.minItemsInCache = 1

	c.Put("test1", 100)
	e, ok := c.lookup["test1"]
	is.True(ok)
	is.Equal(e.lastUsedAge, uint(0))
	is.Equal(e.timesUsed, uint(1))
	is.Equal(e.size, datasize.Bit)
	is.Equal(c.Length(), 1)

	v, ok := c.Get("test1")
	is.True(ok)
	is.Equal(v, int8(100))

	c.Put("test2", 101)
	e, ok = c.lookup["test2"]
	is.True(ok)
	is.Equal(e.lastUsedAge, uint(2))
	is.Equal(e.timesUsed, uint(1))
	is.Equal(e.size, datasize.Bit)
	is.Equal(c.Length(), 2)

	v, ok = c.Get("test2")
	is.True(ok)
	is.Equal(v, int8(101))

	c.Put("test3", 103)
	e, ok = c.lookup["test3"]
	is.True(ok)
	is.Equal(e.lastUsedAge, uint(4))
	is.Equal(e.timesUsed, uint(1))
	is.Equal(e.size, datasize.Bit)
	is.Equal(c.Length(), 3)

	v, ok = c.Get("test3")
	is.True(ok)
	is.Equal(v, int8(103))

	c.Put("test4", 104)
	e, ok = c.lookup["test4"]
	is.True(ok)
	is.Equal(e.lastUsedAge, uint(6))
	is.Equal(e.timesUsed, uint(1))
	is.Equal(e.size, datasize.Bit)
	is.Equal(c.Length(), 3)

	v, ok = c.Get("test4")
	is.True(ok)
	is.Equal(v, int8(104))

	// At this point test1 needs to be evicted.
	_, ok = c.lookup["test1"]
	is.Equal(ok, false)
	_, ok = c.lookup["test2"]
	is.True(ok)
	_, ok = c.lookup["test3"]
	is.True(ok)

	res := c.Cached("test2", func() int8 {
		t.Fail()
		return 0
	})
	is.Equal(res, int8(101))

	res = c.Cached("teserdetst", func() int8 {
		return 69
	})
	is.Equal(res, int8(69))
}
