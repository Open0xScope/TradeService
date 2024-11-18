package handler

import (
	"fmt"
	"sync"
	"time"

	"github.com/Open0xScope/CommuneXService/utils/logger"
	"github.com/sirupsen/logrus"
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

func CheckTradeRateLimit(pubkey string) bool {
	mutex.Lock()
	defer mutex.Unlock()

	now := time.Now()

	accessList, ok := accessMap[pubkey]
	if !ok {
		accessList = []time.Time{}
	}

	// delete more than one minite
	for i := len(accessList) - 1; i >= 0; i-- {
		if now.Sub(accessList[i]) > time.Minute {
			accessList = accessList[i+1:]
			break
		}
	}

	if len(accessList) >= 20 {
		logger.Logrus.WithFields(logrus.Fields{"PubKey": pubkey, "Data": accessList}).Info("CheckRateLimit more than access limit")

		return false
	}

	//update record
	accessList = append(accessList, now)
	accessMap[pubkey] = accessList

	return true
}

func CheckTradeRateLimitDay(pubkey string) bool {
	mutexday.Lock()
	defer mutexday.Unlock()

	now := time.Now()

	accessList, ok := accessDayMap[pubkey]
	if !ok {
		accessList = []time.Time{}
	}

	// delete more than one day
	dat := 24 * time.Hour
	for i := len(accessList) - 1; i >= 0; i-- {
		if now.Sub(accessList[i]) > dat {
			accessList = accessList[i+1:]
			break
		}
	}

	if len(accessList) >= 100 {
		logger.Logrus.WithFields(logrus.Fields{"PubKey": pubkey}).Info("CheckTradeRateLimitDay more than access limit")

		return false
	}

	//update record
	accessList = append(accessList, now)
	accessDayMap[pubkey] = accessList

	return true
}

func CheckTradeTokenRateLimitDay(pubkey, tokenAddr string) bool {
	mutextokenday.Lock()
	defer mutextokenday.Unlock()

	now := time.Now()

	key := fmt.Sprintf("%s%s", pubkey, tokenAddr)

	accessTokenList, ok := accessTokenDayMap[key]
	if !ok {
		accessTokenList = []time.Time{}
	}

	// delete more than one day
	dat := 24 * time.Hour
	for i := len(accessTokenList) - 1; i >= 0; i-- {
		if now.Sub(accessTokenList[i]) > dat {
			accessTokenList = accessTokenList[i+1:]
			break
		}
	}

	if len(accessTokenList) >= 50 {
		logger.Logrus.WithFields(logrus.Fields{"PubKey": pubkey, "TokenAddress": tokenAddr}).Info("CheckTradeTokenRateLimitDay more than access limit")

		return false
	}

	//update record
	accessTokenList = append(accessTokenList, now)
	accessTokenDayMap[key] = accessTokenList

	return true
}

func CheckQueryRateLimit(pubkey string) bool {
	qmutex.Lock()
	defer qmutex.Unlock()

	now := time.Now()

	accessList, ok := accessQueryMap[pubkey]
	if !ok {
		accessList = []time.Time{}
	}

	// delete more than one minite
	for i := len(accessList) - 1; i >= 0; i-- {
		if now.Sub(accessList[i]) > time.Minute {
			accessList = accessList[i+1:]
			break
		}
	}

	if len(accessList) >= 60 {
		logger.Logrus.WithFields(logrus.Fields{"PubKey": pubkey, "Data": accessList}).Info("CheckQueryRateLimit more than access limit")

		return false
	}

	//update record
	accessList = append(accessList, now)
	accessQueryMap[pubkey] = accessList

	return true
}
