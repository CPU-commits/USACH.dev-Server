package stack

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type pubSubClient struct {
	redis      *redis.Client
	subscriber *redis.PubSub
	tasks      map[string]func(payload []byte)
	ctx        context.Context
}

func (rdb *pubSubClient) runSub() {
	for {
		msg, err := rdb.subscriber.ReceiveMessage(rdb.ctx)
		if err != nil {
			continue
		}

		rdb.tasks[msg.Channel]([]byte(msg.Payload))
	}
}

func (rdb *pubSubClient) Sub(event string, handler func(payload []byte)) {
	if rdb.subscriber == nil {
		rdb.subscriber = rdb.redis.Subscribe(rdb.ctx, event)
		go rdb.runSub()
	} else {
		rdb.subscriber.Subscribe(rdb.ctx, event)
	}
	// Add to tasks
	rdb.tasks[event] = handler
}

func (rdb *pubSubClient) Emit(event string, data interface{}) error {
	return rdb.redis.Publish(rdb.ctx, event, data).Err()
}

func NewPubSubClient() *pubSubClient {
	return &pubSubClient{
		redis: rdb,
		ctx:   context.Background(),
		tasks: map[string]func(payload []byte){},
	}
}
