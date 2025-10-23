//go:build integration

package bgg_test

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/kkjdaniel/gogeek/thing"
	"github.com/kkjdaniel/gogeek/user"
	"github.com/ngoldack/dicetrace/package/bgg"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	testredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"golang.org/x/sync/semaphore"
)

var (
	sharedRedisClient *redis.Client
	dbPool            *redisDBPool
)

// redisDBPool manages a pool of Redis database numbers (0-15)
// ensuring only 16 databases are in use at any given time.
type redisDBPool struct {
	sem       *semaphore.Weighted
	available chan int
	mu        sync.Mutex
}

func newRedisDBPool() *redisDBPool {
	pool := &redisDBPool{
		sem:       semaphore.NewWeighted(16),
		available: make(chan int, 16),
	}
	// Initialize with database numbers 0-15
	for i := range 16 {
		pool.available <- i
	}
	return pool
}

// acquire gets a database number from the pool, blocking if all are in use
func (p *redisDBPool) acquire(ctx context.Context) (int, error) {
	// Acquire semaphore weight (blocks if all 16 are in use)
	if err := p.sem.Acquire(ctx, 1); err != nil {
		return 0, err
	}

	// Get a database number from the available pool
	dbNum := <-p.available
	return dbNum, nil
}

// release returns a database number to the pool after flushing it
func (p *redisDBPool) release(dbNum int, client *redis.Client) {
	// Flush the database before returning it to the pool
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = client.FlushDB(ctx).Err()

	// Return the database number to the available pool
	p.available <- dbNum

	// Release the semaphore weight
	p.sem.Release(1)
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Initialize the database pool
	dbPool = newRedisDBPool()

	// Start a single Redis container for all tests
	redisContainer, err := testredis.Run(ctx,
		"redis:7-alpine",
		testredis.WithSnapshotting(10, 1),
		testredis.WithLogLevel(testredis.LogLevelVerbose),
	)
	if err != nil {
		panic("failed to start redis container: " + err.Error())
	}

	// Get connection string
	connStr, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		panic("failed to get redis connection string: " + err.Error())
	}

	// Parse Redis options from connection string
	opts, err := redis.ParseURL(connStr)
	if err != nil {
		panic("failed to parse redis connection string: " + err.Error())
	}

	// Create Redis client
	sharedRedisClient = redis.NewClient(opts)

	// Verify connection
	if err := sharedRedisClient.Ping(ctx).Err(); err != nil {
		panic("failed to ping redis: " + err.Error())
	}

	// Run tests
	code := m.Run()

	// Cleanup
	if sharedRedisClient != nil {
		_ = sharedRedisClient.Close()
	}
	if redisContainer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = redisContainer.Terminate(ctx)
	}

	os.Exit(code)
}

func setupTestRedisClient(t *testing.T) redis.UniversalClient {
	t.Helper()

	// Acquire a database number from the pool (blocks if all 16 are in use)
	ctx := context.Background()
	dbNum, err := dbPool.acquire(ctx)
	require.NoError(t, err, "failed to acquire database from pool")

	// Create a new client pointing to the acquired database
	client := redis.NewClient(&redis.Options{
		Addr: sharedRedisClient.Options().Addr,
		DB:   dbNum,
	})

	// Flush the database to ensure clean state
	err = client.FlushDB(ctx).Err()
	require.NoError(t, err, "failed to flush test database")

	// Cleanup function to flush and release the database back to the pool
	t.Cleanup(func() {
		dbPool.release(dbNum, client)
		_ = client.Close()
	})

	return client
}

func TestRedisBGGCache_Thing_SetAndGet(t *testing.T) {
	t.Parallel()
	client := setupTestRedisClient(t)

	cache := bgg.NewRedisBGGCache(client)
	ctx := context.Background()

	// Create test data
	testThing := &thing.Item{
		ID:   123,
		Type: "boardgame",
	}

	// Test SetThing
	err := cache.SetThing(ctx, testThing)
	require.NoError(t, err, "SetThing should not return an error")

	// Test GetThing
	retrieved, err := cache.GetThing(ctx, 123)
	require.NoError(t, err, "GetThing should not return an error")
	assert.NotNil(t, retrieved, "retrieved thing should not be nil")
	assert.Equal(t, testThing.ID, retrieved.ID, "thing ID should match")
	assert.Equal(t, testThing.Type, retrieved.Type, "thing type should match")
}

func TestRedisBGGCache_Thing_GetMiss(t *testing.T) {
	t.Parallel()
	client := setupTestRedisClient(t)

	cache := bgg.NewRedisBGGCache(client)
	ctx := context.Background()

	// Test GetThing with non-existent ID
	retrieved, err := cache.GetThing(ctx, 999)
	assert.Error(t, err, "GetThing should return an error for cache miss")
	assert.Nil(t, retrieved, "retrieved thing should be nil on cache miss")
	assert.True(t, errors.Is(err, bgg.ErrCacheMiss), "error should be ErrCacheMiss")
}

func TestRedisBGGCache_User_SetAndGet(t *testing.T) {
	t.Parallel()
	client := setupTestRedisClient(t)

	cache := bgg.NewRedisBGGCache(client)
	ctx := context.Background()

	// Create test data
	testUser := &user.User{
		ID:   456,
		Name: "testuser",
	}

	// Test SetUser
	err := cache.SetUser(ctx, testUser)
	require.NoError(t, err, "SetUser should not return an error")

	// Test GetUser
	retrieved, err := cache.GetUser(ctx, "testuser")
	require.NoError(t, err, "GetUser should not return an error")
	assert.NotNil(t, retrieved, "retrieved user should not be nil")
	assert.Equal(t, testUser.ID, retrieved.ID, "user ID should match")
	assert.Equal(t, testUser.Name, retrieved.Name, "user name should match")
}

func TestRedisBGGCache_User_GetMiss(t *testing.T) {
	t.Parallel()
	client := setupTestRedisClient(t)

	cache := bgg.NewRedisBGGCache(client)
	ctx := context.Background()

	// Test GetUser with non-existent username
	retrieved, err := cache.GetUser(ctx, "nonexistentuser")
	assert.Error(t, err, "GetUser should return an error for cache miss")
	assert.Nil(t, retrieved, "retrieved user should be nil on cache miss")
	assert.True(t, errors.Is(err, bgg.ErrCacheMiss), "error should be ErrCacheMiss")
}

func TestRedisBGGCache_Thing_Overwrite(t *testing.T) {
	t.Parallel()
	client := setupTestRedisClient(t)

	cache := bgg.NewRedisBGGCache(client)
	ctx := context.Background()

	// Create initial test data
	thing1 := &thing.Item{
		ID:   789,
		Type: "boardgame",
	}

	// Set initial thing
	err := cache.SetThing(ctx, thing1)
	require.NoError(t, err)

	// Create updated test data with same ID
	thing2 := &thing.Item{
		ID:   789,
		Type: "boardgameexpansion",
	}

	// Overwrite with new data
	err = cache.SetThing(ctx, thing2)
	require.NoError(t, err)

	// Retrieve and verify it's the updated version
	retrieved, err := cache.GetThing(ctx, 789)
	require.NoError(t, err)
	assert.Equal(t, thing2.Type, retrieved.Type, "thing type should be updated")
}

func TestRedisBGGCache_User_Overwrite(t *testing.T) {
	t.Parallel()
	client := setupTestRedisClient(t)

	cache := bgg.NewRedisBGGCache(client)
	ctx := context.Background()

	// Create initial test data
	user1 := &user.User{
		ID:   111,
		Name: "testuser2",
	}

	// Set initial user
	err := cache.SetUser(ctx, user1)
	require.NoError(t, err)

	// Create updated test data with same username
	user2 := &user.User{
		ID:   222,
		Name: "testuser2",
	}

	// Overwrite with new data
	err = cache.SetUser(ctx, user2)
	require.NoError(t, err)

	// Retrieve and verify it's the updated version
	retrieved, err := cache.GetUser(ctx, "testuser2")
	require.NoError(t, err)
	assert.Equal(t, user2.ID, retrieved.ID, "user ID should be updated")
}

func TestRedisBGGCache_ConcurrentAccess(t *testing.T) {
	t.Parallel()
	client := setupTestRedisClient(t)

	cache := bgg.NewRedisBGGCache(client)
	ctx := context.Background()

	// Test concurrent writes and reads
	const numGoroutines = 10

	// Set initial thing
	testThing := &thing.Item{
		ID:   999,
		Type: "boardgame",
	}
	err := cache.SetThing(ctx, testThing)
	require.NoError(t, err)

	// Launch concurrent readers
	type result struct {
		retrieved *thing.Item
		err       error
	}
	results := make(chan result, numGoroutines)

	for range numGoroutines {
		go func() {
			retrieved, err := cache.GetThing(ctx, 999)
			results <- result{retrieved: retrieved, err: err}
		}()
	}

	// Wait for all goroutines to complete and verify results
	for range numGoroutines {
		res := <-results
		assert.NoError(t, res.err)
		assert.NotNil(t, res.retrieved)
		if res.retrieved != nil {
			assert.Equal(t, testThing.ID, res.retrieved.ID)
		}
	}
}

func TestRedisBGGCache_ContextCancellation(t *testing.T) {
	t.Parallel()
	client := setupTestRedisClient(t)

	cache := bgg.NewRedisBGGCache(client)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Try to get a thing with cancelled context
	_, err := cache.GetThing(ctx, 123)
	assert.Error(t, err, "GetThing should fail with cancelled context")

	// Try to set a thing with cancelled context
	testThing := &thing.Item{ID: 123, Type: "boardgame"}
	err = cache.SetThing(ctx, testThing)
	assert.Error(t, err, "SetThing should fail with cancelled context")
}

func TestRedisBGGCache_MultipleThings(t *testing.T) {
	t.Parallel()
	client := setupTestRedisClient(t)

	cache := bgg.NewRedisBGGCache(client)
	ctx := context.Background()

	// Create multiple test things
	things := []*thing.Item{
		{ID: 1, Type: "boardgame"},
		{ID: 2, Type: "boardgame"},
		{ID: 3, Type: "boardgameexpansion"},
	}

	// Store all things
	for _, th := range things {
		err := cache.SetThing(ctx, th)
		require.NoError(t, err)
	}

	// Retrieve and verify all things
	for _, expected := range things {
		retrieved, err := cache.GetThing(ctx, expected.ID)
		require.NoError(t, err)
		assert.Equal(t, expected.ID, retrieved.ID)
		assert.Equal(t, expected.Type, retrieved.Type)
	}
}

func TestRedisBGGCache_MultipleUsers(t *testing.T) {
	t.Parallel()
	client := setupTestRedisClient(t)

	cache := bgg.NewRedisBGGCache(client)
	ctx := context.Background()

	// Create multiple test users
	users := []*user.User{
		{ID: 1, Name: "user1"},
		{ID: 2, Name: "user2"},
		{ID: 3, Name: "user3"},
	}

	// Store all users
	for _, u := range users {
		err := cache.SetUser(ctx, u)
		require.NoError(t, err)
	}

	// Retrieve and verify all users
	for _, expected := range users {
		retrieved, err := cache.GetUser(ctx, expected.Name)
		require.NoError(t, err)
		assert.Equal(t, expected.ID, retrieved.ID)
		assert.Equal(t, expected.Name, retrieved.Name)
	}
}
