package redis

import (
	"sync"

	"github.com/Open0xScope/CommuneXService/config"

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
			Username:     redisConfig.Name,
			Password:     redisConfig.Password,
			DB:           int(redisConfig.DB),
			MinIdleConns: int(redisConfig.MinIdleConns),
		}

		client := redis.NewClient(options)

		redisClient = client
	})
	return redisClient
}
