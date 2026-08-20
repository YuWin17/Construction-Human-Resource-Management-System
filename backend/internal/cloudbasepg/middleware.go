package cloudbasepg

import (
	"context"
	"log/slog"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Synchronizer serializes API requests in the single-instance Cloud Run service
// and commits their durable changes to CloudBase PG after successful responses.
type Synchronizer struct {
	client *Client
	db     *gorm.DB
	logger *slog.Logger
	mu     sync.Mutex
}

func NewSynchronizer(client *Client, db *gorm.DB, logger *slog.Logger) *Synchronizer {
	return &Synchronizer{client: client, db: db, logger: logger}
}

func (s *Synchronizer) Apply(ctx context.Context, before, after Snapshot) error {
	return s.client.ApplyChanges(ctx, before, after)
}

func (s *Synchronizer) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Next()
			return
		}

		s.mu.Lock()
		defer s.mu.Unlock()
		before, err := TakeSnapshot(s.db)
		if err != nil {
			s.logger.Error("snapshot CloudBase working set", "error", err)
			c.AbortWithStatus(503)
			return
		}

		c.Next()
		if c.Writer.Status() >= 400 {
			return
		}
		after, err := TakeSnapshot(s.db)
		if err == nil {
			err = s.client.ApplyChanges(c.Request.Context(), before, after)
		}
		if err != nil {
			// Retain the error in Cloud Run logs without leaking API credentials.
			s.logger.Error("persist CloudBase PG changes", "path", c.Request.URL.Path, "error", err)
			_ = c.Error(err)
		}
	}
}
