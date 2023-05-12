package stack

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type Stack struct {
	redisClient *redis.Client
}

func (s *Stack) Set(key, value string, exp time.Duration) error {
	return s.redisClient.Set(ctx, key, value, exp).Err()
}

func (s *Stack) Get(key string) (string, error) {
	return s.redisClient.Get(ctx, key).Result()
}

func (s *Stack) Delete(keys ...string) error {
	return s.redisClient.Del(ctx, keys...).Err()
}

func NewStack() *Stack {
	return &Stack{
		redisClient: rdb,
	}
}
