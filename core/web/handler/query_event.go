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
	"0x514910771af9ca656af840dff83e8264ecf986ca",
	"0x1f9840a85d5af5bf1d1762f925bdaddc4201f984",
	"0x6982508145454ce325ddbe47a25d4ec3d2311933",
	"0xaea46a60368a7bd060eec7df8cba43b7ef41ad85",
	"0x808507121b80c02388fad14726482e061b8da827",
	"0x9d65ff81a3c488d585bbfb0bfe3c7707c7917f54",
	"0x6e2a43be0b1d33b726f0ca3b8de60b3482b8b050",
	"0xc18360217d8f7ab5e7c516566761ea12ce7f9d72",
	"0xa9b1eb5908cfc3cdf91f9b8b3a74108598009096",
	"0x57e114b691db790c35207b2e685d4a43181e6061",
	"0x6de037ef9ad2725eb40118bb1702ebb27e4aeb24",
	"0x4200000000000000000000000000000000000042",
	"0x912ce59144191c1204e64559fe8253a0e49e6548",
	"0x5a98fcbea516cf06857215779fd812ca3bef1b32",
	"0x62d0a8458ed7719fdaf978fe5929c6d342b0bfce",
	"0x7fc66500c84a76ad7e9c93437bfc5ac33e2ddae9",
	"0xfaba6f8e4a5e8ab82f62fe7c39859fa577269be3",
	"0xc011a73ee8576fb46f5e1c5751ca3b9fe0af2a6f",
	"0x4d224452801aced8b2f0aebe155379bb5d594381",
	"0x5283d291dbcf85356a21ba090e6db59121208b44",
	"0x8290333cef9e6d528dd5618fb97a76f268f3edd4",
	"0x8457ca5040ad67fdebbcc8edce889a335bc0fbfb",
	"0x25f8087ead173b73d6e8b84329989a8eea16cf73",
	"0xaaee1a9723aadb7afa2810263653a34ba2c21c7a",
	"0x6b3595068778dd592e39a122f4f5a5cf09c90fe2",
	"0xc011a73ee8576fb46f5e1c5751ca3b9fe0af2a6f",
	"0xed04915c23f00a313a544955524eb7dbd823143d",
	"0xb528edbef013aff855ac3c50b381f253af13b997",
	"0x64bc2ca1be492be7185faa2c8835d9b824c8a194",
	"0x7a58c0be72be218b41c608b7fe7c5bb630736c71",
	"0x4b9278b94a1112cad404048903b8d343a810b07e",
	"0x72e4f9f808c49a2a61de9c5896298920dc4eeea9",
	"0x14778860e937f509e651192a90589de711fb88a9",
	"0xdef1ca1fb7fbcdc777520aa7f396b4e015f497ab",
	"0xe28b3b32b6c345a34ff64674606124dd5aceca30",
	"0xc00e94cb662c3520282e6f5717214004a7f26888",
	"0x92d6c1e31e14520e676a687f0a93788b716beff5",
	"0x163f8c2467924be0ae7b5347228cabf260318753",
	"0x4b0f1812e5df2a09796481ff14017e6005508003",
	"0x767fe9edc9e0df98e07454847909b5e959d7ca0e",
	"0x7448c7456a97769f6cd04f1e83a4a23ccdc46abd",
	"0x7420b4b9a0110cdc71fb720908340c03f9bc03ec",
	"0x58b6a8a3302369daec383334672404ee733ab239",
	"0x41e5560054824ea6b0732e656e3ad64e20e94e45",
	"0x5faa989af96af85384b8a938c2ede4a7378d9875",
	"0x38e382f74dfb84608f3c1f10187f6bef5951de93",
	"0xc944e90c64b2c07662a292be6244bdf05cda44a7",
	"0x9813037ee2218799597d83d4a5b6f3b6778218d9",
	"0x5b7533812759b45c2b44c19e320ba2cd2681b542",
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
