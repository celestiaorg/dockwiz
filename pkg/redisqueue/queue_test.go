package redisqueue_test

import (
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/celestiaorg/dockwiz/pkg/redisqueue"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnqueueAndDequeueWithMiniRedis(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err, "Error starting miniredis server")
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
		DB:   0,
	})

	queue := redisqueue.NewQueue(rdb, "test_queue")

	// Test Enqueue
	err = queue.Enqueue("item1")
	assert.NoError(t, err, "Error enqueueing item")

	// Test Dequeue
	var item string
	err = queue.Dequeue(&item)
	assert.NoError(t, err, "Error dequeuing item")
	assert.Equal(t, "item1", item, "Dequeued item should match")

	// Test dequeue on an empty queue
	err = queue.Dequeue(&item)
	assert.ErrorIs(t, err, redisqueue.ErrQueueEmpty, "Error should be ErrQueueEmpty on dequeue from an empty queue")
}
