package builder_test

import (
	"testing"
	"time"

	"github.com/celestiaorg/dockwiz/pkg/builder"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCatchLogsHook(t *testing.T) {
	hook := builder.NewCatchLogsHook()

	levels := hook.Levels()
	assert.ElementsMatch(t, logrus.AllLevels, levels, "Levels should match")

	entry := &logrus.Entry{
		Logger:  logrus.New(),
		Level:   logrus.InfoLevel,
		Message: "Test message",
	}
	assert.NoError(t, hook.Fire(entry), "Error should be nil when firing the log entry")

	// Test StreamNewLogs method
	logChan, stop := hook.StreamNewLogs()
	defer stop()

	// Simulate log events
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.AddHook(hook)

	logger.Info("New log message 1")
	logger.Info("New log message 2")

	// Wait for logs to be streamed
	time.Sleep(300 * time.Millisecond)

	logsBuffer := <-logChan

	assert.Contains(t, logsBuffer, "New log message 1", "Log message 1 should be present")
	assert.Contains(t, logsBuffer, "New log message 2", "Log message 2 should be present")
}
