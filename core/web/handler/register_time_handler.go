package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Open0xScope/CommuneXService/core/db"
	"github.com/Open0xScope/CommuneXService/core/model"
	"github.com/Open0xScope/CommuneXService/utils/logger"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func getAllRegistertimes(times string) ([]model.AdsMinerPerformance, error) {
	ctx := context.Background()
	res := make([]model.AdsMinerPerformance, 0)

	err := db.GetDB().NewSelect().Model(&res).Where("register_time >= ?", times).Order("register_time DESC").Limit(50000).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func GetRegisterTime(c *gin.Context) {
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

	starttimeStr := c.Query("starttime")

	logger.Logrus.WithFields(logrus.Fields{"Address": userIdStr, "PubKey": pubKeyStr, "Timestamp": timeStr, "Signature": sigStr, "StartTime": starttimeStr}).Info("GetRegisterTime info")

	rawData := fmt.Sprintf("%s%s%s", userIdStr, pubKeyStr, timeStr)
	err := VerifySign(rawData, pubKeyStr, sigStr)
	if err != nil {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("GetRegisterTime VerifySign failed")
		r.Code = http.StatusInternalServerError
		r.Message = "verify sig failed"
		return
	}

	if ok := CheckQueryRateLimit(pubKeyStr); !ok {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("GetRegisterTime CheckQueryRateLimit failed")
		r.Code = http.StatusTooManyRequests
		r.Message = "access limit exceeded, please try again later"
		return
	}

	isMiner, err := IsMinerOrValidor(userIdStr)
	if err != nil {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("GetRegisterTime validator not registered")
		r.Code = http.StatusInternalServerError
		r.Message = "validator not registered"
		return
	}

	if isMiner {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("GetRegisterTime validator has no access to get all trades")
		r.Code = http.StatusInternalServerError
		r.Message = "validator has no access"
		return
	}

	result, err := getAllRegistertimes(starttimeStr)
	if err != nil {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("GetRegisterTime getAllRegistertime failed")
		r.Code = http.StatusInternalServerError
		r.Message = "get all register time  failed"
		return
	}

	r.Message = "get all register time success"
	r.Data = result
}
