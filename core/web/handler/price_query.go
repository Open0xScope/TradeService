package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Open0xScope/CommuneXService/core/db"
	"github.com/Open0xScope/CommuneXService/core/model"
	"github.com/Open0xScope/CommuneXService/utils/logger"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func getTokenPrice(token string, timestamp int64) (*model.ChainTokenPrice, error) {
	ctx := context.Background()
	var res model.ChainTokenPrice

	mathtime := time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")

	err := db.GetDB().NewSelect().Model(&res).Where("chain in (?) and token_address = ? and pt <= ?", bun.In(ChainList), token, mathtime).Order("pt DESC").Limit(1).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func getLatestPrice(timestr string) ([]model.ChainTokenPrice, error) {
	res := make([]model.ChainTokenPrice, 0)
	if timestr == "" {
		err := db.GetDB().NewSelect().Model(&res).Where("chain in (?) and rn = 1 and token_address in (?)", bun.In(ChainList), bun.In(TokenList)).Scan(context.Background())
		if err != nil {
			return nil, err
		}
	} else {
		ts, err := strconv.ParseInt(timestr, 10, 64)
		if err != nil {
			return nil, err
		}
		mathtime := time.Unix(ts, 0).Format("2006-01-02 15:04:05")

		err = db.GetDB().NewSelect().Model(&res).Where("chain in (?) and token_address in (?) and pt = ?", bun.In(ChainList), bun.In(TokenList), mathtime).Scan(context.Background())
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

func GetLatestPrice(c *gin.Context) {
	r := &Response{
		Code:    http.StatusOK,
		Message: "success",
	}
	defer func(r *Response) {
		c.JSON(http.StatusOK, r)
	}(r)

	userIdStr := c.Query("userId")
	pubKeyStr := c.Query("pubKey")
	timeStr := c.Query("timestamp")
	sigStr := c.Query("sig")

	latestStr := c.Query("latesttime")

	logger.Logrus.WithFields(logrus.Fields{"MinerID": userIdStr, "PubKey": pubKeyStr, "Timestamp": timeStr, "Signature": sigStr, "LatestTime": latestStr}).Info("GetLatestPrice info")

	rawData := fmt.Sprintf("%s%s%s", userIdStr, pubKeyStr, timeStr)
	err := VerifySign(rawData, pubKeyStr, sigStr)
	if err != nil {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("GetLatestPrice VerifySign failed")
		r.Code = http.StatusInternalServerError
		r.Message = "verify sig failed"
		return
	}

	if ok := CheckQueryRateLimit(pubKeyStr); !ok {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("GetLatestPrice CheckQueryRateLimit failed")
		r.Code = http.StatusTooManyRequests
		r.Message = "access limit exceeded, please try again later"
		return
	}

	result, err := getLatestPrice(latestStr)
	if err != nil {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("GetLatestPrice getLatestPrice failed")
		r.Code = http.StatusInternalServerError
		r.Message = "get latest price failed"
		return
	}

	logger.Logrus.WithFields(logrus.Fields{"Data": result}).Debug("GetLatestPrice getLatestPrice info")

	r.Message = "get latest price success"
	r.Data = result
}
