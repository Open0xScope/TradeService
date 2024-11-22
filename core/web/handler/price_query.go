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
	var res model.ChainTokenPrice
	mathtime := time.Unix(timestamp, 0).UTC().Format("2006-01-02 15:04:05")

	// Build the subquery
	subquery := db.GetDB().NewSelect().
		Table("crawler_ods.ods_crawler_coingecko_trade_token_price").
		Column("*").
		ColumnExpr("row_number() OVER (PARTITION BY token_address ORDER BY pt DESC) AS rn").
		Where("chain IN (?)", bun.In(ChainList)).
		Where("token_address = ?", token).
		Where("pt <= ?", mathtime)

	// Build the main query
	err := db.GetDB().NewSelect().
		TableExpr("(?) AS a", subquery).
		Where("rn = 1").
		Scan(context.Background(), &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func getLatestPrice(timestr string) ([]model.ChainTokenPrice, error) {
	res := make([]model.ChainTokenPrice, 0)
	mathtime := ""
	if timestr == "" {
		mathtime = "2100-01-02 15:04:05"
	} else {
		ts, err := strconv.ParseInt(timestr, 10, 64)
		if err != nil {
			return nil, err
		}
		mathtime = time.Unix(ts, 0).UTC().Format("2006-01-02 15:04:05")
	}

	// Build the subquery
	subquery := db.GetDB().NewSelect().
		Table("crawler_ods.ods_crawler_coingecko_trade_token_price").
		Column("*").
		ColumnExpr("row_number() OVER (PARTITION BY token_address ORDER BY pt DESC) AS rn").
		Where("chain IN (?)", bun.In(ChainList)).
		Where("token_address IN (?)", bun.In(TokenList)).
		Where("pt <= ?", mathtime)

	// Build the main query
	err := db.GetDB().NewSelect().
		TableExpr("(?) AS a", subquery).
		Where("rn = 1").
		Scan(context.Background(), &res)
	if err != nil {
		return nil, err
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

	err = CheckQueryRateLimit(pubKeyStr)
	if err != nil {
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
