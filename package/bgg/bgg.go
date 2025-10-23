package bgg

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/kkjdaniel/gogeek/thing"
	"github.com/kkjdaniel/gogeek/user"
	"tailscale.com/util/singleflight"
)

type BGGService interface {
	FetchThing(ctx context.Context, id int) (*thing.Item, error)
	FetchUser(ctx context.Context, username string) (*user.User, error)
}

type bggServiceImpl struct {
	cache   BGGCache
	sfUser  singleflight.Group[string, *user.User]
	sfThing singleflight.Group[int, *thing.Item]
}

func NewBGGService(cache BGGCache) BGGService {
	return &bggServiceImpl{
		cache:   cache,
		sfUser:  singleflight.Group[string, *user.User]{},
		sfThing: singleflight.Group[int, *thing.Item]{},
	}
}

func (s *bggServiceImpl) FetchThing(ctx context.Context, id int) (*thing.Item, error) {
	t, err := s.cache.GetThing(ctx, id)
	if err != nil && !errors.Is(err, ErrCacheMiss) {
		return nil, err
	}

	if t != nil {
		// found in cache
		slog.DebugContext(ctx, "thing found in cache; returning", slog.Any("thing", t))
		return t, nil
	}

	res := <-s.sfThing.DoChanContext(ctx, id, func(ctx context.Context) (*thing.Item, error) {
		items, err := thing.Query([]int{id})
		if err != nil {
			return nil, err
		}
		if len(items.Items) == 0 {
			return nil, fmt.Errorf("thing with id '%d' not found", id)
		}

		t = &items.Items[0]
		slog.DebugContext(ctx, "thing fetched from BGG API", slog.Any("thing", t))
		err = s.cache.SetThing(ctx, t)
		if err != nil {
			return nil, err
		}

		return t, nil
	})
	if res.Err != nil {
		return nil, fmt.Errorf("failed to fetch thing with id '%d': %w", id, res.Err)
	}
	slog.DebugContext(ctx, "thing fetched from BGG API via singleflight", slog.Any("result", res))

	t = res.Val

	return t, nil
}

func (s *bggServiceImpl) FetchUser(ctx context.Context, username string) (*user.User, error) {
	u, err := s.cache.GetUser(ctx, username)
	if err != nil && !errors.Is(err, ErrCacheMiss) {
		return nil, fmt.Errorf("failed to get user from cache: %w", err)
	}

	if u != nil {
		// found in cache
		slog.DebugContext(ctx, "user found in cache; returning", slog.Any("user", u))
		return u, nil
	}

	res := <-s.sfUser.DoChanContext(ctx, username, func(ctx context.Context) (*user.User, error) {
		usr, err := user.Query(username)
		if err != nil {
			return nil, fmt.Errorf("failed to query user from BGG API: %w", err)
		}

		slog.DebugContext(ctx, "user fetched from BGG API", slog.Any("user", usr))
		err = s.cache.SetUser(ctx, usr)
		if err != nil {
			return nil, fmt.Errorf("failed to set user in cache: %w", err)
		}

		return usr, nil
	})
	if res.Err != nil {
		return nil, fmt.Errorf("failed to fetch user '%s': %w", username, res.Err)
	}
	slog.DebugContext(ctx, "user fetched from BGG API via singleflight", slog.Any("result", res))

	u = res.Val

	return u, nil
}
