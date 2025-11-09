package bloomfilter

import (
	"context"
	"hash/fnv"

	"github.com/redis/go-redis/v9"
)

// Filter defines the interface for a Bloom filter.
type Filter interface {
	Add(ctx context.Context, value string) error
	Test(ctx context.Context, value string) (bool, error)
	AddMulti(ctx context.Context, values []string) error
}

// redisBloomFilter implements the Filter interface using Redis bitmaps.
type redisBloomFilter struct {
	client *redis.Client
	key    string
	size   uint // m
	hashes uint // k
}

// NewRedisBloomFilter creates a new Redis-backed Bloom filter.
func NewRedisBloomFilter(client *redis.Client, key string, size, hashes uint) Filter {
	return &redisBloomFilter{
		client: client,
		key:    key,
		size:   size,
		hashes: hashes,
	}
}

// locations generates k hash values for a given value.
// It uses a double hashing technique to generate multiple hash values from two base hashes.
func (bf *redisBloomFilter) locations(value string) []uint {
	locations := make([]uint, bf.hashes)
	h1 := fnv.New64a()
	h1.Write([]byte(value))
	hash1 := h1.Sum64()

	h2 := fnv.New64()
	h2.Write([]byte(value))
	hash2 := h2.Sum64()

	for i := uint(0); i < bf.hashes; i++ {
		locations[i] = uint((hash1 + uint64(i)*hash2) % uint64(bf.size))
	}
	return locations
}

// Add adds a value to the Bloom filter.
func (bf *redisBloomFilter) Add(ctx context.Context, value string) error {
	locations := bf.locations(value)
	pipe := bf.client.Pipeline()
	for _, loc := range locations {
		pipe.SetBit(ctx, bf.key, int64(loc), 1)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// AddMulti adds multiple values to the Bloom filter.
func (bf *redisBloomFilter) AddMulti(ctx context.Context, values []string) error {
	pipe := bf.client.Pipeline()
	for _, value := range values {
		locations := bf.locations(value)
		for _, loc := range locations {
			pipe.SetBit(ctx, bf.key, int64(loc), 1)
		}
	}
	_, err := pipe.Exec(ctx)
	return err
}

// Test checks if a value is possibly in the Bloom filter.
// It returns true if the value might be in the set, and false if it is definitely not.
func (bf *redisBloomFilter) Test(ctx context.Context, value string) (bool, error) {
	locations := bf.locations(value)
	pipe := bf.client.Pipeline()
	results := make([]*redis.IntCmd, len(locations))
	for i, loc := range locations {
		results[i] = pipe.GetBit(ctx, bf.key, int64(loc))
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	for _, res := range results {
		bit, err := res.Result()
		if err != nil {
			return false, err
		}
		if bit == 0 {
			return false, nil // Definitely not in the set
		}
	}

	return true, nil // Possibly in the set
}

