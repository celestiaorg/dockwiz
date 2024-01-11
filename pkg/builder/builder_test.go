package builder_test

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/celestiaorg/dockwiz/pkg/builder"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestBuildStatusMethods(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err, "Error should be nil when starting miniredis server")
	defer mr.Close()

	// Create a Redis client pointing to the miniredis server
	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
		DB:   0,
	})

	logger, err := zap.NewDevelopment()
	require.NoError(t, err, "Error should be nil when creating logger")
	b := builder.NewBuilder(rdb, logger)

	// Test SetBuildStatus
	data := builder.BuildStatusData{
		Status:   1,
		EndTime:  time.Now(),
		ErrorMsg: "Test error",
		Logs:     "Test log",
	}
	err = b.SetBuildStatus("testImage", data)
	require.NoError(t, err, "Error should be nil when setting build status")

	// Test UpdateBuildStatus
	newData := builder.BuildStatusData{
		Status:   2,
		EndTime:  time.Now().Add(5 * time.Second),
		ErrorMsg: "Updated error",
		Logs:     "Updated log",
	}
	err = b.UpdateBuildStatus("testImage", newData)
	require.NoError(t, err, "Error should be nil when updating build status")

	// Test GetBuildStatus
	result, err := b.GetBuildStatus("testImage")
	require.NoError(t, err, "Error should be nil when getting build status")

	assert.True(t, newData.StartTime.Equal(result.StartTime), "Start time should be equal")
	assert.True(t, newData.EndTime.Equal(result.EndTime), "End time should be equal")
	assert.Equal(t, newData.Status, result.Status, "Status should be equal")
	assert.Equal(t, newData.ErrorMsg, result.ErrorMsg, "Error message should be equal")
	// UpdateBuildStatus appends logs, so we need to add the old log to the new log
	assert.Equal(t, data.Logs+newData.Logs, result.Logs, "Log should be equal")
}

func TestAddToBuildQueue(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err, "Error should be nil when starting miniredis server")
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
		DB:   0,
	})

	logger, err := zap.NewDevelopment()
	require.NoError(t, err, "Error should be nil when creating logger")
	b := builder.NewBuilder(rdb, logger)

	opts := builder.BuilderOptions{
		Image: builder.ImageOptions{
			Prefix:      "prefix",
			Name:        "test-image",
			Tag:         "test-tag",
			Destination: "https://test-registry",
		},
		Git: builder.GitOptions{
			URL: "github.com/test-username/test-repo",
		},
	}

	result, err := b.AddToBuildQueue(opts)
	require.NoError(t, err, "Error should be nil when adding to build queue")
	assert.Equal(t, opts.Image.Name, result.ImageName, "Image name should match")
	assert.Equal(t, opts.Image.Tag, result.ImageTag, "Image tag should match")

	var qOpts builder.BuilderOptions
	err = b.Queue.Dequeue(&qOpts)
	require.NoError(t, err, "Error should be nil when dequeuing from the queue")
	assert.Equal(t, opts.Image, qOpts.Image, "Image should match")
	assert.Equal(t, opts.Git.URL, qOpts.Git.URL, "Git URL should match")
}

func TestStartBuilder(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err, "Error should be nil when starting miniredis server")
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
		DB:   0,
	})

	logger, err := zap.NewDevelopment()
	require.NoError(t, err, "Error should be nil when creating logger")
	b := builder.NewBuilder(rdb, logger)

	opts := builder.BuilderOptions{
		Git: builder.GitOptions{
			URL: "github.com/test-username/test-repo",
		},
	}

	err = b.SetBuildStatus(opts.Image.Name, builder.BuildStatusData{Status: builder.StatusPending})
	require.NoError(t, err, "Error should be nil when setting build status")

	err = b.Queue.Enqueue(opts)
	require.NoError(t, err, "Error should be nil when adding build to the queue")

	// Start the builder worker
	b.Start()
	defer b.Close()

	// Sleep to simulate the builder working
	time.Sleep(2 * time.Second)

	// To verify the outcome, check the build status after the worker runs
	bd, err := b.GetBuildStatus(opts.Image.Name)
	require.NoError(t, err, "Error should be nil when getting build status")

	assert.NotEqual(t, builder.StatusPending, bd.Status, "Build status should not be pending")
}
