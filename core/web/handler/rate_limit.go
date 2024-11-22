package handler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Open0xScope/CommuneXService/core/redis"
)

var (
	accessMap = make(map[string][]time.Time)
	mutex     sync.Mutex

	accessDayMap = make(map[string][]time.Time)
	mutexday     sync.Mutex

	accessQueryMap = make(map[string][]time.Time)
	qmutex         sync.Mutex

	accessTokenDayMap = make(map[string][]time.Time)
	mutextokenday     sync.Mutex
)

func chaeckKeyExp(key string, limit int64, exp time.Duration) error {
	ctx := context.Background()
	// Increment the count
	count, err := redis.GetRedisInst().Incr(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to increment key: %v", err)
	}

	if count == 1 {
		err = redis.GetRedisInst().Expire(ctx, key, exp).Err()
		if err != nil {
			return fmt.Errorf("failed to expire key: %v", err)
		}
	}

	if count > int64(limit) {
		return fmt.Errorf("current:maximum %v : %v", count, limit)
	}
	return nil
}

func CheckTradeRateLimit(pubkey string) error {
	redisKey := fmt.Sprintf("trade_min_rate_limit:%s", pubkey)

	err := chaeckKeyExp(redisKey, 20, time.Minute)
	if err != nil {
		return fmt.Errorf("CheckTradeRateLimit,%s %v", pubkey, err)
	}

	return nil
}

func CheckTradeRateLimitDay(pubkey string) error {
	redisKey := fmt.Sprintf("trade_day_rate_limit:%s", pubkey)

	err := chaeckKeyExp(redisKey, 100, 24*time.Hour)
	if err != nil {
		return fmt.Errorf("CheckTradeRateLimitDay,%s %v", pubkey, err)
	}

	return nil
}

func CheckTradeTokenRateLimitDay(pubkey, tokenAddr string) error {
	redisKey := fmt.Sprintf("trade_token_day_rate_limit:%s:%s", pubkey, tokenAddr)

	err := chaeckKeyExp(redisKey, 50, 24*time.Hour)
	if err != nil {
		return fmt.Errorf("CheckTradeTokenRateLimitDay,%s:%s %v", pubkey, tokenAddr, err)
	}

	return nil
}

func CheckQueryRateLimit(pubkey string) error {
	redisKey := fmt.Sprintf("trade_query_min_rate_limit:%s", pubkey)

	err := chaeckKeyExp(redisKey, 60, time.Minute)
	if err != nil {
		return fmt.Errorf("CheckQueryRateLimit,%s %v", pubkey, err)
	}

	return nil
}
