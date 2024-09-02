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
)

func getUserTrades(userId string) ([]model.AdsTokenTrade, error) {
	ctx := context.Background()
	res := make([]model.AdsTokenTrade, 0)
	err := db.GetDB().NewSelect().Model(&res).Where("miner_id = ?", userId).Order("timestamp DESC").Limit(1000).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func getTotalCount() (int, error) {
	ctx := context.Background()

	totalCount, err := db.GetDB().NewSelect().Model(&model.AdsTokenTrade{}).Where("status > 0").Count(ctx)
	if err != nil {
		return 0, err
	}

	return totalCount, nil
}

func getAllTrades(times, pageStr, limitStr string) ([]model.AdsTokenTrade, error) {
	ctx := context.Background()
	res := make([]model.AdsTokenTrade, 0)

	last7daytime := int64(0)

	if times == "" {
		last7daytime = time.Now().UTC().Add(-14 * 24 * time.Hour).Unix()
	} else {
		s, err := strconv.ParseInt(times, 10, 64)
		if err != nil {
			return nil, err
		}

		last7daytime = s
	}

	page, _ := strconv.ParseInt(pageStr, 10, 64)

	limit, _ := strconv.ParseInt(limitStr, 10, 64)
	offset := (page - 1) * limit

	err := db.GetDB().NewSelect().Model(&res).Where("status > 0 and timestamp >= ?", last7daytime).Order("timestamp DESC").Limit(int(limit)).Offset(int(offset)).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func GetUserTraddes(c *gin.Context) {
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

	logger.Logrus.WithFields(logrus.Fields{"MinerID": userIdStr, "PubKey": pubKeyStr, "Timestamp": timeStr, "Signature": sigStr}).Info("GetUserTraddes info")

	rawData := fmt.Sprintf("%s%s%s", userIdStr, pubKeyStr, timeStr)
	err := VerifySign(rawData, pubKeyStr, sigStr)
	if err != nil {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("GetUserTraddes VerifySign failed")
		r.Code = http.StatusInternalServerError
		r.Message = "verify sig failed"
		return
	}

	if ok := CheckQueryRateLimit(pubKeyStr); !ok {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("GetUserTraddes CheckQueryRateLimit failed")
		r.Code = http.StatusTooManyRequests
		r.Message = "access limit exceeded, please try again later"
		return
	}

	result, err := getUserTrades(userIdStr)
	if err != nil {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("GetUserTraddes getUserTrades failed")
		r.Code = http.StatusInternalServerError
		r.Message = "get user trades failed"
		return
	}

	r.Message = "get user trades success"
	r.Data = result
}

func GetAllTraddes(c *gin.Context) {
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
	page := c.Query("page")
	limit := c.Query("limit")

	tradetimeStr := c.Query("tradetime")

	logger.Logrus.WithFields(logrus.Fields{"MinerID": userIdStr, "PubKey": pubKeyStr, "Timestamp": timeStr, "Signature": sigStr, "TradeTime": tradetimeStr, "Page": page, "Limit": limit}).Info("GetAllTraddes info")

	rawData := fmt.Sprintf("%s%s%s", userIdStr, pubKeyStr, timeStr)
	err := VerifySign(rawData, pubKeyStr, sigStr)
	if err != nil {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("GetAllTraddes VerifySign failed")
		r.Code = http.StatusInternalServerError
		r.Message = "verify sig failed"
		return
	}

	if ok := CheckQueryRateLimit(pubKeyStr); !ok {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("GetAllTraddes CheckQueryRateLimit failed")
		r.Code = http.StatusTooManyRequests
		r.Message = "access limit exceeded, please try again later"
		return
	}

	isMiner, err := IsMinerOrValidor(userIdStr)
	if err != nil {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("GetAllTraddes validator not registered")
		r.Code = http.StatusInternalServerError
		r.Message = "validator not registered"
		return
	}

	if isMiner {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("GetAllTraddes validator has no access to get all trades")
		r.Code = http.StatusInternalServerError
		r.Message = "validator has no access"
		return
	}

	totalCount, err := getTotalCount()
	if err != nil {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("GetAllTraddes getTotalCount failed")
		r.Code = http.StatusInternalServerError
		r.Message = "get total count failed"
		return
	}

	result, err := getAllTrades(tradetimeStr, page, limit)
	if err != nil {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("GetAllTraddes getAllTrades failed")
		r.Code = http.StatusInternalServerError
		r.Message = "get all trades failed"
		return
	}

	type TradesResult struct {
		TotalCount int                   `json:"total_count"`
		TradeList  []model.AdsTokenTrade `json:"trade_list"`
	}

	resData := TradesResult{
		TotalCount: totalCount,
		TradeList:  result,
	}

	r.Message = "get all trades success"
	r.Data = resData
}
