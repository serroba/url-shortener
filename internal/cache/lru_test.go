package cache_test

import (
	"sync"
	"testing"

	"github.com/serroba/web-demo-go/internal/cache"
	"github.com/serroba/web-demo-go/internal/shortener"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newShortURL(code, url string) *shortener.ShortURL {
	return &shortener.ShortURL{
		Code:        shortener.Code(code),
		OriginalURL: url,
	}
}

func TestLRU_BasicOperations(t *testing.T) {
	t.Run("get returns false for missing key", func(t *testing.T) {
		c := cache.New(10)

		val, ok := c.Get("missing")

		assert.False(t, ok)
		assert.Nil(t, val)
	})

	t.Run("set and get returns value", func(t *testing.T) {
		c := cache.New(10)
		url := newShortURL("abc", "https://example.com")

		c.Set("abc", url)
		val, ok := c.Get("abc")

		require.True(t, ok)
		assert.Equal(t, url, val)
	})

	t.Run("set updates existing key", func(t *testing.T) {
		c := cache.New(10)
		url1 := newShortURL("abc", "https://example1.com")
		url2 := newShortURL("abc", "https://example2.com")

		c.Set("abc", url1)
		c.Set("abc", url2)
		val, ok := c.Get("abc")

		require.True(t, ok)
		assert.Equal(t, "https://example2.com", val.OriginalURL)
	})

	t.Run("len returns correct count", func(t *testing.T) {
		c := cache.New(10)

		assert.Equal(t, 0, c.Len())

		c.Set("a", newShortURL("a", "https://a.com"))
		assert.Equal(t, 1, c.Len())

		c.Set("b", newShortURL("b", "https://b.com"))
		assert.Equal(t, 2, c.Len())
	})
}

func TestLRU_Eviction(t *testing.T) {
	t.Run("evicts when capacity exceeded", func(t *testing.T) {
		c := cache.New(2)

		c.Set("a", newShortURL("a", "https://a.com"))
		c.Set("b", newShortURL("b", "https://b.com"))
		c.Set("c", newShortURL("c", "https://c.com")) // should evict "a"

		_, ok := c.Get("a")
		assert.False(t, ok, "a should be evicted")

		_, ok = c.Get("b")
		assert.True(t, ok, "b should still exist")

		_, ok = c.Get("c")
		assert.True(t, ok, "c should exist")

		assert.Equal(t, 2, c.Len())
	})

	t.Run("evicts least recently used", func(t *testing.T) {
		c := cache.New(2)

		c.Set("a", newShortURL("a", "https://a.com"))
		c.Set("b", newShortURL("b", "https://b.com"))

		// Access "a" to make it recently used
		c.Get("a")

		// Add "c" - should evict "b" (least recently used)
		c.Set("c", newShortURL("c", "https://c.com"))

		_, ok := c.Get("a")
		assert.True(t, ok, "a should still exist (was accessed)")

		_, ok = c.Get("b")
		assert.False(t, ok, "b should be evicted (LRU)")

		_, ok = c.Get("c")
		assert.True(t, ok, "c should exist")
	})

	t.Run("updating existing key does not evict", func(t *testing.T) {
		c := cache.New(2)

		c.Set("a", newShortURL("a", "https://a.com"))
		c.Set("b", newShortURL("b", "https://b.com"))

		// Update "a" - should not cause eviction
		c.Set("a", newShortURL("a", "https://updated.com"))

		assert.Equal(t, 2, c.Len())

		val, ok := c.Get("a")
		require.True(t, ok)
		assert.Equal(t, "https://updated.com", val.OriginalURL)
	})
}

func TestLRU_Ordering(t *testing.T) {
	t.Run("get moves item to front", func(t *testing.T) {
		c := cache.New(3)

		c.Set("a", newShortURL("a", "https://a.com"))
		c.Set("b", newShortURL("b", "https://b.com"))
		c.Set("c", newShortURL("c", "https://c.com"))

		// Access order: a (oldest), then b, then c (newest)
		// Now access "a" to make it most recent
		c.Get("a")

		// Add "d" - should evict "b" (now the LRU)
		c.Set("d", newShortURL("d", "https://d.com"))

		_, ok := c.Get("a")
		assert.True(t, ok, "a should exist (was accessed, moved to front)")

		_, ok = c.Get("b")
		assert.False(t, ok, "b should be evicted (LRU after a was accessed)")
	})

	t.Run("set moves existing item to front", func(t *testing.T) {
		c := cache.New(3)

		c.Set("a", newShortURL("a", "https://a.com"))
		c.Set("b", newShortURL("b", "https://b.com"))
		c.Set("c", newShortURL("c", "https://c.com"))

		// Update "a" to move it to front
		c.Set("a", newShortURL("a", "https://a-updated.com"))

		// Add "d" - should evict "b" (now the LRU)
		c.Set("d", newShortURL("d", "https://d.com"))

		_, ok := c.Get("a")
		assert.True(t, ok, "a should exist (was updated, moved to front)")

		_, ok = c.Get("b")
		assert.False(t, ok, "b should be evicted (LRU after a was updated)")
	})
}

func TestLRU_Concurrent(t *testing.T) {
	t.Run("concurrent access is safe", func(t *testing.T) {
		c := cache.New(100)

		var wg sync.WaitGroup

		// Writers
		for i := range 10 {
			wg.Add(1)

			go func(id int) {
				defer wg.Done()

				for j := range 100 {
					key := string(rune('a' + (id*100+j)%26))
					c.Set(key, newShortURL(key, "https://example.com"))
				}
			}(i)
		}

		// Readers
		for range 10 {
			wg.Add(1)

			go func() {
				defer wg.Done()

				for range 100 {
					c.Get("a")
					c.Len()
				}
			}()
		}

		wg.Wait()

		// Should complete without race conditions
		assert.LessOrEqual(t, c.Len(), 100)
	})
}
