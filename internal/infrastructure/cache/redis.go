package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/vitao/geolocation-tracker/pkg/config"
	"github.com/vitao/geolocation-tracker/pkg/logger"
)

// Redis representa o cliente Redis para cache
type Redis struct {
	client *redis.Client
	logger logger.Logger
}

// NewRedis cria uma nova instância do cliente Redis
func NewRedis(cfg *config.Config, logger logger.Logger) (*Redis, error) {
	// Criar cliente Redis
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password:     "", // Sem senha por enquanto
		DB:           0,  // DB padrão
		PoolSize:     10,
		MinIdleConns: 2,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// Testar conexão
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis connection established",
		"host", cfg.Redis.Host,
		"port", cfg.Redis.Port,
	)

	return &Redis{
		client: client,
		logger: logger,
	}, nil
}

// Close fecha a conexão com Redis
func (r *Redis) Close() error {
	return r.client.Close()
}

// Health verifica a saúde da conexão Redis
func (r *Redis) Health(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Set armazena um valor no cache
func (r *Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	// Serializar valor para JSON
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	// Armazenar no Redis
	if err := r.client.Set(ctx, key, data, expiration).Err(); err != nil {
		r.logger.Error("Failed to set cache",
			"key", key,
			"error", err.Error(),
		)
		return fmt.Errorf("failed to set cache: %w", err)
	}

	r.logger.Debug("Cache set successfully",
		"key", key,
		"expiration", expiration.String(),
	)

	return nil
}

// Get recupera um valor do cache
func (r *Redis) Get(ctx context.Context, key string, dest interface{}) error {
	// Buscar valor no Redis
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("cache miss: key not found")
		}
		r.logger.Error("Failed to get cache",
			"key", key,
			"error", err.Error(),
		)
		return fmt.Errorf("failed to get cache: %w", err)
	}

	// Deserializar JSON
	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	r.logger.Debug("Cache hit",
		"key", key,
	)

	return nil
}

// Delete remove um valor do cache
func (r *Redis) Delete(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		r.logger.Error("Failed to delete cache",
			"key", key,
			"error", err.Error(),
		)
		return fmt.Errorf("failed to delete cache: %w", err)
	}

	r.logger.Debug("Cache deleted",
		"key", key,
	)

	return nil
}

// Exists verifica se uma chave existe no cache
func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check cache existence: %w", err)
	}

	return result > 0, nil
}

// CacheUserPosition armazena a posição atual de um usuário no cache
func (r *Redis) CacheUserPosition(ctx context.Context, userID string, position interface{}) error {
	key := fmt.Sprintf("user:position:%s", userID)
	expiration := 5 * time.Minute // Cache por 5 minutos

	return r.Set(ctx, key, position, expiration)
}

// GetCachedUserPosition recupera a posição atual de um usuário do cache
func (r *Redis) GetCachedUserPosition(ctx context.Context, userID string, dest interface{}) error {
	key := fmt.Sprintf("user:position:%s", userID)
	return r.Get(ctx, key, dest)
}

// CacheNearbyUsers armazena resultado de busca por proximidade
func (r *Redis) CacheNearbyUsers(ctx context.Context, lat, lng, radius float64, users interface{}) error {
	key := fmt.Sprintf("nearby:%.6f:%.6f:%.0f", lat, lng, radius)
	expiration := 2 * time.Minute // Cache por 2 minutos (dados mais dinâmicos)

	return r.Set(ctx, key, users, expiration)
}

// GetCachedNearbyUsers recupera resultado de busca por proximidade do cache
func (r *Redis) GetCachedNearbyUsers(ctx context.Context, lat, lng, radius float64, dest interface{}) error {
	key := fmt.Sprintf("nearby:%.6f:%.6f:%.0f", lat, lng, radius)
	return r.Get(ctx, key, dest)
}

// LogStats registra estatísticas do Redis
func (r *Redis) LogStats() {
	stats := r.client.PoolStats()
	r.logger.Info("Redis connection stats",
		"hits", stats.Hits,
		"misses", stats.Misses,
		"timeouts", stats.Timeouts,
		"total_conns", stats.TotalConns,
		"idle_conns", stats.IdleConns,
		"stale_conns", stats.StaleConns,
	)
}

// Client retorna o cliente Redis para uso em outras partes do sistema
func (r *Redis) Client() *redis.Client {
	return r.client
}
