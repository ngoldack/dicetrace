package bgg_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kkjdaniel/gogeek/thing"
	"github.com/kkjdaniel/gogeek/user"
	"github.com/ngoldack/dicetrace/package/bgg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCache implements bgg.BGGCache for unit testing
type mockCache struct {
	things      map[int]*thing.Item
	users       map[string]*user.User
	getThingErr error
	setThingErr error
	getUserErr  error
	setUserErr  error
}

func newMockCache() *mockCache {
	return &mockCache{
		things: make(map[int]*thing.Item),
		users:  make(map[string]*user.User),
	}
}

func (m *mockCache) GetThing(ctx context.Context, id int) (*thing.Item, error) {
	if m.getThingErr != nil {
		return nil, m.getThingErr
	}
	if item, ok := m.things[id]; ok {
		return item, nil
	}
	return nil, nil
}

func (m *mockCache) SetThing(ctx context.Context, item *thing.Item) error {
	if m.setThingErr != nil {
		return m.setThingErr
	}
	m.things[item.ID] = item
	return nil
}

func (m *mockCache) GetUser(ctx context.Context, username string) (*user.User, error) {
	if m.getUserErr != nil {
		return nil, m.getUserErr
	}
	if usr, ok := m.users[username]; ok {
		return usr, nil
	}
	return nil, nil
}

func (m *mockCache) SetUser(ctx context.Context, usr *user.User) error {
	if m.setUserErr != nil {
		return m.setUserErr
	}
	m.users[usr.Name] = usr
	return nil
}

func TestBGGService_FetchThing_CacheHit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cache := newMockCache()
	service := bgg.NewBGGService(cache)

	// Pre-populate cache with a thing
	expectedItem := &thing.Item{ID: 42, Type: "boardgame"}
	cache.things[42] = expectedItem

	// Fetch should return cached item
	result, err := service.FetchThing(ctx, 42)
	require.NoError(t, err)
	assert.Equal(t, expectedItem.ID, result.ID)
	assert.Equal(t, expectedItem.Type, result.Type)
}

func TestBGGService_FetchThing_CacheError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cache := newMockCache()
	service := bgg.NewBGGService(cache)

	// Simulate cache error (but not a cache miss)
	testErr := errors.New("redis connection failed")
	cache.getThingErr = testErr

	// Should return the cache error
	result, err := service.FetchThing(ctx, 42)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, testErr, err)
}

func TestBGGService_FetchUser_CacheHit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cache := newMockCache()
	service := bgg.NewBGGService(cache)

	// Pre-populate cache with a user
	expectedUser := &user.User{Name: "testuser", ID: 123}
	cache.users["testuser"] = expectedUser

	// Fetch should return cached user
	result, err := service.FetchUser(ctx, "testuser")
	require.NoError(t, err)
	assert.Equal(t, expectedUser.Name, result.Name)
	assert.Equal(t, expectedUser.ID, result.ID)
}

func TestBGGService_FetchUser_CacheError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cache := newMockCache()
	service := bgg.NewBGGService(cache)

	// Simulate cache error (but not a cache miss)
	testErr := errors.New("redis connection failed")
	cache.getUserErr = testErr

	// Should return the cache error
	result, err := service.FetchUser(ctx, "testuser")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorContains(t, err, "failed to get user from cache")
}

func TestBGGService_NewBGGService(t *testing.T) {
	t.Parallel()
	cache := newMockCache()
	service := bgg.NewBGGService(cache)
	assert.NotNil(t, service)
}

// Note: Testing cache miss scenarios with the real BGG API requires integration tests
// The current implementation calls thing.Query() and user.Query() which are external API calls
// For those scenarios, see cache_integration_test.go
