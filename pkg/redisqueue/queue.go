package redisqueue

import (
	"errors"
	"fmt"

	"github.com/go-redis/redis"
)

type Queue struct {
	client *redis.Client
	name   string
}

var (
	ErrQueueEmpty = errors.New("queue is empty")
)

func NewQueue(client *redis.Client, name string) *Queue {
	return &Queue{
		client: client,
		name:   name,
	}
}

func (q *Queue) Enqueue(item interface{}) error {
	tx := q.client.TxPipeline()
	if err := tx.RPush(q.name, item).Err(); err != nil {
		return fmt.Errorf("enqueue error: %v", err)
	}

	if _, err := tx.Exec(); err != nil {
		return fmt.Errorf("enqueue error: %v", err)
	}
	return nil
}

func (q *Queue) Dequeue(data interface{}) error {
	tx := q.client.TxPipeline()
	if err := tx.LPop(q.name).Err(); err != nil {
		return fmt.Errorf("dequeue error `tx.LPop`: %v", err)
	}
	cmds, err := tx.Exec()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("dequeue error `tx.Exec`: %v", err)
	}

	if len(cmds) != 1 {
		return fmt.Errorf("dequeue error: expected 1 command, got %d", len(cmds))
	}

	if err := cmds[0].(*redis.StringCmd).Scan(data); err != nil {
		if err == redis.Nil {
			return ErrQueueEmpty
		}
		return fmt.Errorf("error converting result: %v", err)
	}
	return nil
}
