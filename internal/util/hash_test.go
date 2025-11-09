package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	password := "mysecretpassword"
	hash, err := HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
}

func TestCheckPasswordHash(t *testing.T) {
	password := "mysecretpassword"
	hash, err := HashPassword(password)
	assert.NoError(t, err)

	assert.True(t, CheckPasswordHash(password, hash))
	assert.False(t, CheckPasswordHash("wrongpassword", hash))
}

func BenchmarkHashPassword(b *testing.B) {
	password := "a_very_long_and_secure_password_for_benchmarking"
	for i := 0; i < b.N; i++ {
		_, _ = HashPassword(password)
	}
}

func BenchmarkCheckPasswordHash(b *testing.B) {
	password := "a_very_long_and_secure_password_for_benchmarking"
	hash, _ := HashPassword(password)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CheckPasswordHash(password, hash)
	}
}
```
```go
