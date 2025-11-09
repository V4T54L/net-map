package bloomfilter

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func setupTestRedis(t testing.TB) *redis.Client {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	return client
}

func TestRedisBloomFilter(t *testing.T) {
	client := setupTestRedis(t)
	ctx := context.Background()
	bf := NewRedisBloomFilter(client, "test_bloom", 1000, 3)

	t.Run("Add and Test", func(t *testing.T) {
		// Add an item
		err := bf.Add(ctx, "hello")
		require.NoError(t, err)

		// Test for the added item (should be true)
		exists, err := bf.Test(ctx, "hello")
		require.NoError(t, err)
		require.True(t, exists, "item 'hello' should exist in the filter")

		// Test for a non-existent item (should be false)
		exists, err = bf.Test(ctx, "world")
		require.NoError(t, err)
		require.False(t, exists, "item 'world' should not exist in the filter")
	})

	t.Run("AddMulti", func(t *testing.T) {
		items := []string{"apple", "banana", "cherry"}
		err := bf.AddMulti(ctx, items)
		require.NoError(t, err)

		for _, item := range items {
			exists, err := bf.Test(ctx, item)
			require.NoError(t, err)
			require.True(t, exists, "item '%s' should exist after AddMulti", item)
		}

		exists, err := bf.Test(ctx, "grape")
		require.NoError(t, err)
		require.False(t, exists, "item 'grape' should not exist")
	})
}

func BenchmarkBloomFilter_Add(b *testing.B) {
	client := setupTestRedis(b)
	ctx := context.Background()
	bf := NewRedisBloomFilter(client, "bench_bloom_add", 100000, 5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.Add(ctx, "item"+string(rune(i)))
	}
}

func BenchmarkBloomFilter_Test(b *testing.B) {
	client := setupTestRedis(b)
	ctx := context.Background()
	bf := NewRedisBloomFilter(client, "bench_bloom_test", 100000, 5)

	// Pre-populate the filter
	for i := 0; i < 10000; i++ {
		bf.Add(ctx, "item"+string(rune(i)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.Test(ctx, "item"+string(rune(i)))
	}
}
```
