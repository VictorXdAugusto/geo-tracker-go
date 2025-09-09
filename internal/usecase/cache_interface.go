package usecase

import (
	"context"
	"time"
)

// CacheInterface define os métodos do cache que os use cases precisam
type CacheInterface interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error

	// Helper methods
	CacheUserPosition(ctx context.Context, userID string, position interface{}) error
	GetCachedUserPosition(ctx context.Context, userID string, dest interface{}) error
	CacheNearbyUsers(ctx context.Context, lat, lng, radius float64, users interface{}) error
	GetCachedNearbyUsers(ctx context.Context, lat, lng, radius float64, dest interface{}) error
	CacheUserHistory(ctx context.Context, userID string, limit int, history interface{}) error
	GetCachedUserHistory(ctx context.Context, userID string, limit int, dest interface{}) error
	InvalidateUserCaches(ctx context.Context, userID string) error
}
