package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	feedTTL   = 5 * time.Minute
	digestTTL = 10 * time.Minute
)

type Cache struct {
	rdb *redis.Client
}

func New(redisURL string) (*Cache, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	return &Cache{rdb: redis.NewClient(opt)}, nil
}

func (c *Cache) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

func (c *Cache) Close() error {
	return c.rdb.Close()
}

func (c *Cache) Get(ctx context.Context, key string, dst any) (bool, error) {
	val, err := c.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("redis get %s: %w", key, err)
	}
	if err := json.Unmarshal(val, dst); err != nil {
		return false, fmt.Errorf("unmarshal: %w", err)
	}
	return true, nil
}

func (c *Cache) Set(ctx context.Context, key string, val any, ttl time.Duration) error {
	b, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return c.rdb.Set(ctx, key, b, ttl).Err()
}

func (c *Cache) FeedKey(sourceID, limit, offset int, search string) string {
	return fmt.Sprintf("feed:src=%d:lim=%d:off=%d:q=%s", sourceID, limit, offset, search)
}

func (c *Cache) DigestKey() string {
	return "digest:" + time.Now().Format("2006-01-02")
}

func (c *Cache) FeedTTL() time.Duration  { return feedTTL }
func (c *Cache) DigestTTL() time.Duration { return digestTTL }
