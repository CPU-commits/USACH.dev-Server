package stack

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var rdb *redis.Client

func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     settingsData.REDIS_URI,
		Password: settingsData.REDIS_PASS,
		DB:       settingsData.REDIS_DB,
	})
}
