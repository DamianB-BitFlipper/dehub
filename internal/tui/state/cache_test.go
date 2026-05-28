package state

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCacheEvictsLeastRecentlyUsedEntry(t *testing.T) {
	c := NewCache[int](2)
	c.Put("a", 1)
	c.Put("b", 2)

	_, ok := c.Get("a")
	require.True(t, ok)

	c.Put("c", 3)

	_, ok = c.Get("b")
	require.False(t, ok)
	value, ok := c.Get("a")
	require.True(t, ok)
	require.Equal(t, 1, value)
	value, ok = c.Get("c")
	require.True(t, ok)
	require.Equal(t, 3, value)
}

func TestCacheUpdatesExistingEntry(t *testing.T) {
	c := NewCache[int](1)
	c.Put("a", 1)
	c.Put("a", 2)

	value, ok := c.Get("a")
	require.True(t, ok)
	require.Equal(t, 2, value)
	require.Equal(t, 1, c.Len())
}
