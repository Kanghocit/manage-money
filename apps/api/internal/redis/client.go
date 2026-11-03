package redisstore

import (
	"context"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Store struct {
	client *redis.Client
}

func Connect(url string) (*Store, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		if strings.HasPrefix(url, "redis://") {
			return nil, err
		}
		opt = &redis.Options{Addr: url}
	}
	client := redis.NewClient(opt)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return &Store{client: client}, nil
}

func (s *Store) Client() *redis.Client {
	return s.client
}

func (s *Store) BlacklistRefresh(ctx context.Context, jti string, ttl time.Duration) error {
	return s.client.Set(ctx, "bl:refresh:"+jti, "1", ttl).Err()
}

func (s *Store) IsRefreshBlacklisted(ctx context.Context, jti string) (bool, error) {
	n, err := s.client.Exists(ctx, "bl:refresh:"+jti).Result()
	return n > 0, err
}

func (s *Store) CacheDashboard(ctx context.Context, userID string, payload []byte, ttl time.Duration) error {
	return s.client.Set(ctx, "cache:dashboard:"+userID, payload, ttl).Err()
}

func (s *Store) GetDashboardCache(ctx context.Context, userID string) ([]byte, error) {
	return s.client.Get(ctx, "cache:dashboard:"+userID).Bytes()
}

func (s *Store) InvalidateDashboard(ctx context.Context, userID string) error {
	return s.client.Del(ctx, "cache:dashboard:"+userID).Err()
}
