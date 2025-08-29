package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/skr1ms/mosaic/config"
	"github.com/skr1ms/mosaic/pkg/middleware"
)

var logger = middleware.NewLogger()

// NewRedisClient creates a new Redis client with the configuration
func NewRedisClient(cfg *config.Config) (*redis.Client, error) {
	opt, err := redis.ParseURL(cfg.RedisConfig.URL)
	if err != nil {
		logger.GetZerologLogger().Error().Err(err).Msg("Failed to parse Redis URL")
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	// Additional connection parameters
	opt.MaxRetries = 3
	opt.DialTimeout = 5 * time.Second
	opt.ReadTimeout = 3 * time.Second
	opt.WriteTimeout = 3 * time.Second
	opt.PoolSize = 10
	opt.MinIdleConns = 5

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logger.GetZerologLogger().Error().Err(err).Msg("Failed to connect to Redis")
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.GetZerologLogger().Info().Str("redis_url", cfg.RedisConfig.URL).Msg("Successfully connected to Redis")
	return client, nil
}
