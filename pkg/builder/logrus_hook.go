package builder

import (
	"bytes"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	logFetchInterval = 100 * time.Millisecond
)

type CatchLogsHook struct {
	Logs         *bytes.Buffer
	lastReadPos  int64 // Track the last read position
	lastReadLock sync.Mutex
}

var _ logrus.Hook = (*CatchLogsHook)(nil)

func NewCatchLogsHook() *CatchLogsHook {
	return &CatchLogsHook{
		Logs:        bytes.NewBuffer([]byte{}),
		lastReadPos: 0,
	}
}

func (h *CatchLogsHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire is triggered when a log event is fired
func (h *CatchLogsHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}

	// Write log entry to the buffer
	h.lastReadLock.Lock()
	defer h.lastReadLock.Unlock()
	if _, err := h.Logs.WriteString(line); err != nil {
		return err
	}

	return nil
}

// StreamNewLogs continuously streams new logs since the last read position
func (h *CatchLogsHook) StreamNewLogs() (chan string, func()) {
	logChan := make(chan string)
	ticker := time.NewTicker(logFetchInterval)

	go func() {
		for range ticker.C {
			h.lastReadLock.Lock()
			newLogs := h.Logs.Bytes()[h.lastReadPos:]
			h.lastReadPos = int64(h.Logs.Len())
			h.lastReadLock.Unlock()

			if len(newLogs) > 0 {
				logChan <- string(newLogs)
			}
		}
	}()

	return logChan, func() {
		// Wait for the last logs to be streamed
		time.Sleep(500 * time.Millisecond)
		ticker.Stop()
		close(logChan)
	}
}
