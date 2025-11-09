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
	size   uint
	hashes uint
}

// NewRedisBloomFilter creates a new Bloom filter backed by Redis.
// size: the size of the bit array (m).
// hashes: the number of hash functions (k).
func NewRedisBloomFilter(client *redis.Client, key string, size, hashes uint) Filter {
	return &redisBloomFilter{
		client: client,
		key:    key,
		size:   size,
		hashes: hashes,
	}
}

func (bf *redisBloomFilter) locations(value string) []uint {
	locations := make([]uint, bf.hashes)
	h1 := fnv.New64a()
	h1.Write([]byte(value))
	hash1 := h1.Sum64()

	h2 := fnv.New64()
	h2.Write([]byte(value))
	hash2 := h2.Sum64()

	for i := uint(0); i < bf.hashes; i++ {
		// Double hashing to generate k hash values
		loc := (hash1 + uint64(i)*hash2) % uint64(bf.size)
		locations[i] = uint(loc)
	}
	return locations
}

func (bf *redisBloomFilter) Add(ctx context.Context, value string) error {
	pipe := bf.client.Pipeline()
	for _, loc := range bf.locations(value) {
		pipe.SetBit(ctx, bf.key, int64(loc), 1)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (bf *redisBloomFilter) AddMulti(ctx context.Context, values []string) error {
	pipe := bf.client.Pipeline()
	for _, value := range values {
		for _, loc := range bf.locations(value) {
			pipe.SetBit(ctx, bf.key, int64(loc), 1)
		}
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (bf *redisBloomFilter) Test(ctx context.Context, value string) (bool, error) {
	pipe := bf.client.Pipeline()
	results := make([]*redis.IntCmd, bf.hashes)
	for i, loc := range bf.locations(value) {
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

