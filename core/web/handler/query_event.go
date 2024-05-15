package handler

import (
	"context"
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

var TokenList = []string{
	"0x514910771af9ca656af840dff83e8264ecf986ca", "0x1f9840a85d5af5bf1d1762f925bdaddc4201f984",
	"0x6982508145454ce325ddbe47a25d4ec3d2311933", "0xaea46a60368a7bd060eec7df8cba43b7ef41ad85",
	"0x808507121b80c02388fad14726482e061b8da827", "0x9d65ff81a3c488d585bbfb0bfe3c7707c7917f54",
	"0x6e2a43be0b1d33b726f0ca3b8de60b3482b8b050", "0xc18360217d8f7ab5e7c516566761ea12ce7f9d72",
	"0xa9b1eb5908cfc3cdf91f9b8b3a74108598009096", "0x57e114b691db790c35207b2e685d4a43181e6061",
}

func getAllEvents(startStr, endStr string) ([]model.AdsTokenEvents, error) {
	start := ""
	end := ""

	if startStr == "" {
		start = time.Now().UTC().Add(-90 * 24 * time.Hour).Format("2006-01-02 15:04:05")[:13]
	} else {
		s, err := strconv.ParseInt(startStr, 10, 64)
		if err != nil {
			return nil, err
		}

		start = time.Unix(s, 0).UTC().Format("2006-01-02 15:04:05")[:13]
	}

	if endStr == "" {
		end = time.Now().UTC().Format("2006-01-02 15:04:05")[:13]
	} else {
		s, err := strconv.ParseInt(endStr, 10, 64)
		if err != nil {
			return nil, err
		}

		end = time.Unix(s, 0).UTC().Format("2006-01-02 15:04:05")[:13]
	}

	ctx := context.Background()
	res := make([]model.AdsTokenEvents, 0)

	err := db.GetDB().NewSelect().Model(&res).Where("chain = ? and token_address in (?) and pt BETWEEN ? AND ?", "eth", bun.In(TokenList), start, end).Order("pt DESC").Scan(ctx)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func GetAllEvents(c *gin.Context) {
	r := &Response{
		Code:    http.StatusOK,
		Message: "success",
	}
	defer func(r *Response) {
		c.JSON(http.StatusOK, r)
	}(r)

	startStr := c.Query("start")
	endStr := c.Query("end")

	logger.Logrus.WithFields(logrus.Fields{"Start": startStr, "End": endStr}).Info("GetAllEvents info")

	result, err := getAllEvents(startStr, endStr)
	if err != nil {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("GetAllTraddes getAllEvents failed")
		r.Code = http.StatusInternalServerError
		r.Message = "get all events failed"
		return
	}

	logger.Logrus.WithFields(logrus.Fields{"Data": result}).Info("GetAllTraddes getAllEvents info")

	r.Message = "get all events success"
	r.Data = result
}
