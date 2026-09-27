package cache

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// TaskListTTL is the required 60s cache lifetime for GET /api/tasks.
	TaskListTTL = 60 * time.Second
	// TaskListPrefix namespaces all list-query cache entries.
	TaskListPrefix = "tasks:list:"
)

// Cache is a minimal interface so service/handler can be unit tested
// with miniredis or an in-memory fake.
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	DeleteByPattern(ctx context.Context, pattern string) error
	Ping(ctx context.Context) error
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(addr, password string, db int) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &RedisCache{client: client}, nil
}

func NewRedisCacheFromClient(c *redis.Client) *RedisCache {
	return &RedisCache{client: c}
}

func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// DeleteByPattern deletes keys matching pattern via SCAN (safe for prod).
func (r *RedisCache) DeleteByPattern(ctx context.Context, pattern string) error {
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return r.client.Del(ctx, keys...).Err()
}

func (r *RedisCache) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// BuildTasksKey builds a deterministic cache key that includes query parameters.
// Requirement: "Cache key must include query parameters."
func BuildTasksKey(params url.Values) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	sb.WriteString(TaskListPrefix)
	first := true
	for _, k := range keys {
		vals := params[k]
		sort.Strings(vals)
		for _, v := range vals {
			if !first {
				sb.WriteString("&")
			}
			sb.WriteString(url.QueryEscape(k))
			sb.WriteString("=")
			sb.WriteString(url.QueryEscape(v))
			first = false
		}
	}
	if first {
		// No params — still a stable key.
		sb.WriteString("all")
	}
	return sb.String()
}

// BuildTasksKeyFromMap is a convenience for tests/callers without url.Values.
func BuildTasksKeyFromMap(m map[string]string) string {
	v := url.Values{}
	for k, val := range m {
		v.Set(k, val)
	}
	return BuildTasksKey(v)
}

// InvalidateTaskLists removes all cached GET /api/tasks responses.
func InvalidateTaskLists(ctx context.Context, c Cache) error {
	if c == nil {
		return nil
	}
	return c.DeleteByPattern(ctx, TaskListPrefix+"*")
}

var _ = fmt.Sprintf // keep fmt import if unused in future edits
