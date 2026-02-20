package infrastructure

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/common-lib/utils"
)

func InitRedis(ctx context.Context) (*redis.Client, error) {
	redisAddr := utils.MustGetEnv("REDIS_ADDR")

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,

		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		logging.Errorf(ctx, "could not connect to redis at: %s", redisAddr)
		return nil, err
	}

	logging.Infof(ctx, "connected to redis at: %s", redisAddr)
	return rdb, nil
}
