package infrastructure

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
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
		log.Print("could not connect to redis at: ", redisAddr)
		return nil, err
	}

	log.Print("connected to redis at: ", redisAddr)
	return rdb, nil
}
