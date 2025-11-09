package cache

import (
	"context"
	"internal-dns/internal/domain"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
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

func TestDnsCacheRedis(t *testing.T) {
	client := setupTestRedis(t)
	cache := NewDNSRecordCache(client)
	ctx := context.Background()

	record := &domain.DNSRecord{
		ID:         1,
		UserID:     1,
		DomainName: "test.local",
		Type:       domain.A,
		Value:      "192.168.1.1",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	t.Run("Get miss", func(t *testing.T) {
		res, err := cache.Get(ctx, "nonexistent.local")
		assert.Nil(t, res)
		assert.ErrorIs(t, err, ErrCacheMiss)
	})

	t.Run("Set and Get hit", func(t *testing.T) {
		err := cache.Set(ctx, record)
		require.NoError(t, err)

		res, err := cache.Get(ctx, "test.local")
		require.NoError(t, err)
		require.NotNil(t, res)

		assert.Equal(t, record.ID, res.ID)
		assert.Equal(t, record.DomainName, res.DomainName)
		assert.Equal(t, record.Value, res.Value)
	})

	t.Run("Delete", func(t *testing.T) {
		err := cache.Set(ctx, record)
		require.NoError(t, err)

		err = cache.Delete(ctx, "test.local")
		require.NoError(t, err)

		res, err := cache.Get(ctx, "test.local")
		assert.Nil(t, res)
		assert.ErrorIs(t, err, ErrCacheMiss)
	})
}

