```go
package repository

import (
	"context"
	"internal-dns/internal/domain"
)

type AuditLogRepository interface {
	Create(ctx context.Context, log *domain.AuditLog) error
}
```
