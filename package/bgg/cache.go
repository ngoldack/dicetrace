package bgg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/kkjdaniel/gogeek/thing"
	"github.com/kkjdaniel/gogeek/user"
	"github.com/redis/go-redis/v9"
)

const (
	thingCacheTTL    = 24 * time.Hour
	thingCachePrefix = "bgg:thing:"

	userCacheTTL    = 24 * time.Hour
	userCachePrefix = "bgg:user:"
)

var ErrCacheMiss = fmt.Errorf("cache miss")

// BGGCache defines caching operations for BGG data.
type BGGCache interface {
	GetThing(ctx context.Context, id int) (*thing.Item, error)
	SetThing(ctx context.Context, item *thing.Item) error

	GetUser(ctx context.Context, username string) (*user.User, error)
	SetUser(ctx context.Context, user *user.User) error
}

type RedisBGGCache struct {
	rc redis.UniversalClient
}

func NewRedisBGGCache(rc redis.UniversalClient) *RedisBGGCache {
	return &RedisBGGCache{
		rc: rc,
	}
}

func (c *RedisBGGCache) GetThing(ctx context.Context, id int) (*thing.Item, error) {
	key := generateThingCacheKey(id)
	data, err := c.rc.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("thing '%d' cache miss: %w", id, errors.Join(ErrCacheMiss, err))
	} else if err != nil {
		return nil, err
	}

	var item thing.Item
	err = json.Unmarshal(data, &item)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal thing cache data: %w", err)
	}

	return &item, nil
}

func (c *RedisBGGCache) SetThing(ctx context.Context, item *thing.Item) error {
	key := generateThingCacheKey(item.ID)
	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal thing cache data: %w", err)
	}

	err = c.rc.Set(ctx, key, data, thingCacheTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to set thing cache: %w", err)
	}

	return nil
}

func generateThingCacheKey(id int) string {
	return fmt.Sprintf("%s%d", thingCachePrefix, id)
}

func (c *RedisBGGCache) GetUser(ctx context.Context, username string) (*user.User, error) {
	key := generateUserCacheKey(username)
	data, err := c.rc.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("user '%s' cache miss: %w", username, errors.Join(ErrCacheMiss, err))
	} else if err != nil {
		return nil, err
	}

	var usr user.User
	err = json.Unmarshal(data, &usr)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user cache data: %w", err)
	}

	return &usr, nil
}

func (c *RedisBGGCache) SetUser(ctx context.Context, usr *user.User) error {
	key := generateUserCacheKey(usr.Name)
	data, err := json.Marshal(usr)
	if err != nil {
		return fmt.Errorf("failed to marshal user cache data: %w", err)
	}

	err = c.rc.Set(ctx, key, data, userCacheTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to set user cache: %w", err)
	}

	return nil
}

func generateUserCacheKey(username string) string {
	return userCachePrefix + username
}
