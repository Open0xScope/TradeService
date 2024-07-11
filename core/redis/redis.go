package redis

import (
	"context"
	"os"
	"sync"

	"github.com/Open0xScope/CommuneXService/config"
	"github.com/Open0xScope/CommuneXService/utils/logger"
	"github.com/sirupsen/logrus"

	redis "github.com/go-redis/redis/v8"
)

const Nil = redis.Nil

// one DB one client
var redisClient *redis.Client
var once sync.Once

func InitRedis() error {
	redisClient = GetRedisInst()
	return nil
}

func GetRedisInst() *redis.Client {
	once.Do(func() {
		redisConfig := config.GetRedisConfig()
		options := &redis.Options{
			Addr:         redisConfig.Host,
			Password:     "",
			DB:           int(redisConfig.DB),
			MinIdleConns: int(redisConfig.MinIdleConns),
			PoolSize:     10,
		}

		client := redis.NewClient(options)

		// Ping the Redis server
		pong, err := client.Ping(context.Background()).Result()
		if err != nil {
			logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("connect redis failed")
			os.Exit(0)
		}

		logger.Logrus.WithFields(logrus.Fields{"PongMsg": pong}).Info("connect redis success")

		redisClient = client
	})
	return redisClient
}
