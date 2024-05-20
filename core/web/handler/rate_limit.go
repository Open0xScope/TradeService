package handler

import (
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

	if len(accessList) >= 200 {
		logger.Logrus.WithFields(logrus.Fields{"PubKey": pubkey, "Data": accessList}).Info("CheckTradeRateLimitDay more than access limit")

		return false
	}

	//update record
	accessList = append(accessList, now)
	accessDayMap[pubkey] = accessList

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
