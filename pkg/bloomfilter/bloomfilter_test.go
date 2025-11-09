package bloomfilter

import (
	"context"
	"fmt"
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
	bf := NewRedisBloomFilter(client, "test_bloom", 1000, 4)
	ctx := context.Background()

	t.Run("Add and Test", func(t *testing.T) {
		item1 := "hello"
		item2 := "world"

		// Test before adding
		exists, err := bf.Test(ctx, item1)
		require.NoError(t, err)
		require.False(t, exists)

		// Add item1
		err = bf.Add(ctx, item1)
		require.NoError(t, err)

		// Test after adding
		exists, err = bf.Test(ctx, item1)
		require.NoError(t, err)
		require.True(t, exists)

		// Test non-existent item2
		exists, err = bf.Test(ctx, item2)
		require.NoError(t, err)
		require.False(t, exists)
	})

	t.Run("AddMulti", func(t *testing.T) {
		items := []string{"apple", "banana", "cherry"}
		err := bf.AddMulti(ctx, items)
		require.NoError(t, err)

		for _, item := range items {
			exists, err := bf.Test(ctx, item)
			require.NoError(t, err)
			require.True(t, exists, "item %s should exist", item)
		}

		exists, err := bf.Test(ctx, "grape")
		require.NoError(t, err)
		require.False(t, exists)
	})
}

func BenchmarkBloomFilter_Add(b *testing.B) {
	client := setupTestRedis(b)
	bf := NewRedisBloomFilter(client, "bench_bloom_add", 100000, 4)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		item := fmt.Sprintf("item-%d", i)
		_ = bf.Add(ctx, item)
	}
}

func BenchmarkBloomFilter_Test(b *testing.B) {
	client := setupTestRedis(b)
	bf := NewRedisBloomFilter(client, "bench_bloom_test", 100000, 4)
	ctx := context.Background()

	// Pre-populate the filter
	for i := 0; i < 10000; i++ {
		item := fmt.Sprintf("item-%d", i)
		_ = bf.Add(ctx, item)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		item := fmt.Sprintf("item-%d", i)
		_, _ = bf.Test(ctx, item)
	}
}

