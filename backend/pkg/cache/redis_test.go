package cache

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestCache(t *testing.T) (*RedisCache, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return NewRedisCacheFromClient(client), mr
}

func TestBuildTasksKeyIncludesQueryParams(t *testing.T) {
	a := url.Values{}
	a.Set("status", "todo")
	a.Set("page", "1")
	a.Set("keyword", "fix bug")

	b := url.Values{}
	b.Set("keyword", "fix bug")
	b.Set("page", "1")
	b.Set("status", "todo")

	// Order-independent.
	assert.Equal(t, BuildTasksKey(a), BuildTasksKey(b))

	c := url.Values{}
	c.Set("status", "done")
	c.Set("page", "1")
	c.Set("keyword", "fix bug")
	assert.NotEqual(t, BuildTasksKey(a), BuildTasksKey(c))

	assert.Contains(t, BuildTasksKey(a), "status=todo")
	assert.Contains(t, BuildTasksKey(a), "keyword=fix+bug")
}

func TestSetGetWithTTL(t *testing.T) {
	c, mr := newTestCache(t)
	ctx := context.Background()
	require.NoError(t, c.Set(ctx, "tasks:list:test", `{"data":[]}`, TaskListTTL))
	val, err := c.Get(ctx, "tasks:list:test")
	require.NoError(t, err)
	assert.Equal(t, `{"data":[]}`, val)
	assert.Equal(t, 60*time.Second, mr.TTL("tasks:list:test"))
}

func TestCacheInvalidation(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "tasks:list:status=todo", "a", time.Minute))
	require.NoError(t, c.Set(ctx, "tasks:list:status=done", "b", time.Minute))
	require.NoError(t, c.Set(ctx, "other:key", "keep", time.Minute))

	require.NoError(t, InvalidateTaskLists(ctx, c))

	_, err := c.Get(ctx, "tasks:list:status=todo")
	assert.Error(t, err)
	_, err = c.Get(ctx, "tasks:list:status=done")
	assert.Error(t, err)
	// Non-matching keys survive.
	kept, err := c.Get(ctx, "other:key")
	require.NoError(t, err)
	assert.Equal(t, "keep", kept)
}

func TestInvalidateNilCacheIsNoop(t *testing.T) {
	assert.NoError(t, InvalidateTaskLists(context.Background(), nil))
}
