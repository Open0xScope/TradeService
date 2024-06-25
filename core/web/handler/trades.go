package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/Open0xScope/CommuneXService/core/db"
	"github.com/Open0xScope/CommuneXService/core/model"
	"github.com/Open0xScope/CommuneXService/utils/logger"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type InCreateTrade struct {
	MinerID         string `json:"miner_id"`
	PubKey          string `json:"pub_key"`
	Nonce           int64  `json:"nonce"`
	Token           string `json:"token"`
	PositionManager string `json:"position_manager"`
	Direction       int    `json:"direction"`
	Timestamp       int64  `json:"timestamp"`
	Leverage        string `json:"leverage"`
	Signature       string `json:"signature"`
}

func isDivisible(a, b float64) bool {
	if b == 0 {
		return false
	}
	quotient := a / b
	return math.Abs(quotient-math.Round(quotient)) < 1e-9
}

func insertTrade(txs *model.AdsTokenTrade) error {
	ctx := context.Background()
	sqlRes, err := db.GetDB().NewInsert().Model(txs).Exec(ctx)
	if err != nil {
		return err
	}

	num, err := sqlRes.RowsAffected()
	if err != nil {
		return err
	}

	if num < 0 {
		return errors.New("insert empty item")
	}

	return nil
}

func getLatestTrade(userId, tokenAddr string) (*model.AdsTokenTrade, error) {
	ctx := context.Background()
	var res model.AdsTokenTrade
	err := db.GetDB().NewSelect().Model(&res).Where("miner_id = ? and token = ?", userId, tokenAddr).Order("timestamp DESC").Limit(1).Scan(ctx)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func IsMinerOrValidor(minerid string) (bool, error) {
	//true is miner,and false is validator when error is not null
	ctx := context.Background()
	var res model.AdsMinerWhitelist
	err := db.GetDB().NewSelect().Model(&res).Where("address = ?", minerid).Scan(ctx)
	if err == sql.ErrNoRows {
		return false, errors.New("miner is not in whitelist")
	}

	if err != nil {
		return false, err
	}

	if res.Stake > 1000000000000 {
		return false, nil
	}

	if res.Status < 1 {
		return false, errors.New("miner is not invalided")
	}

	return true, nil
}

func checkTradeValid(latestTrade, newTrade *model.AdsTokenTrade, leverageStr string) (string, error) {
	_, err := IsMinerOrValidor(newTrade.MinerID)
	if err != nil {
		return "miner not registered", err
	}

	err = CheckAddress(newTrade.PubKey, newTrade.MinerID)
	if err != nil {
		return "address and key not match", err
	}

	msg := fmt.Sprintf("%s%s%d%s%s%d%d%s", newTrade.MinerID, newTrade.PubKey, newTrade.Nonce, newTrade.TokenAddress, newTrade.PositionManager, newTrade.Direction, newTrade.Timestamp, leverageStr)
	err = VerifySign(msg, newTrade.PubKey, newTrade.Signature)
	if err != nil {
		return "sign error", err
	}

	if ok := CheckTradeRateLimit(newTrade.PubKey); !ok {
		return "access limit exceeded, please try again later", err
	}

	if ok := CheckTradeRateLimitDay(newTrade.PubKey); !ok {
		return "access day limit exceeded, please try again later", err
	}

	leverage := float64(0.0)
	if leverageStr != "" {
		leverage, err = strconv.ParseFloat(leverageStr, 64)
		if err != nil {
			return "parse leverage failed", err
		}

		if leverage < 0.2 || leverage > 5 {
			return "the leverage is not in the range", errors.New("the leverage is not in the range")
		}

		if !isDivisible(leverage, 0.1) {
			return "the leverage is not an integer multiple of the basic unit", errors.New("the leverage is not an integer multiple of the basic unit")
		}
	} else {
		leverage = float64(1.0)
	}

	newTrade.Leverage = leverage

	if latestTrade != nil {
		if newTrade.Nonce == latestTrade.Nonce {
			return "invalid nonce", errors.New("trade has invalid nonce")
		}

		if newTrade.Timestamp <= latestTrade.Timestamp {
			return "creation timestamp of latest trade error", errors.New("the timestamp of trade is invalid")
		}

		if newTrade.PositionManager == "close" && latestTrade.PositionManager != "open" {
			return "close position manager error", errors.New("close trade is invalid")
		}

		if newTrade.PositionManager == "open" && latestTrade.PositionManager != "close" {
			return "open position manager error", errors.New("open trade is invalid")
		}
	}

	if newTrade.PositionManager != "open" && newTrade.PositionManager != "close" {
		return "position manager invalid", errors.New("position manager is invalid")
	}

	if newTrade.Timestamp > time.Now().Unix() {
		return "creation timestamp more than now", errors.New("trade has invalid timestamp")
	}

	last10min := time.Now().Add(-time.Minute).Unix()
	if newTrade.Timestamp < last10min {
		return "creation timestamp is old", errors.New("trade has invalid timestamp")
	}

	return "check trade success", nil
}

func updatePrice4H(latestTrade, newTrade *model.AdsTokenTrade) error {
	if newTrade.PositionManager != "close" && latestTrade.PositionManager != "open" {
		return nil
	}

	inval := newTrade.Timestamp - latestTrade.Timestamp - 14400
	if inval > 0 {
		return nil
	}

	//update trade 4h price
	latestTrade.TradePrice4H = newTrade.TradePrice

	_, err := db.GetDB().NewUpdate().Model(latestTrade).Set("price_4h = ?", newTrade.TradePrice).Where("miner_id = ? and token = ? and nonce = ?", latestTrade.MinerID, latestTrade.TokenAddress, latestTrade.Nonce).Exec(context.Background())
	if err != nil {
		return fmt.Errorf("update trade 4h price,%v", err)
	}

	return nil
}

func CreateTradde(c *gin.Context) {
	r := &Response{
		Code:    http.StatusOK,
		Message: "success",
	}
	defer func(r *Response) {
		c.JSON(http.StatusOK, r)
	}(r)

	var in = InCreateTrade{}
	err := c.ShouldBind(&in)
	if err != nil {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("CreateTradde parse parmeter failed")
		r.Code = http.StatusBadRequest
		r.Message = "invalid input parameters"
		return
	}

	//after check , insert db
	tradePrice, err := getTokenPrice(in.Token, in.Timestamp)
	if err != nil {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("CreateTradde getTokenPrice failed")
		r.Code = http.StatusBadRequest
		r.Message = "get token price failed"
		return
	}

	newTrade := &model.AdsTokenTrade{
		MinerID:         in.MinerID,
		PubKey:          in.PubKey,
		Nonce:           in.Nonce,
		TokenAddress:    in.Token,
		PositionManager: in.PositionManager,
		Direction:       in.Direction,
		Timestamp:       in.Timestamp,
		TradePrice:      tradePrice.Price,
		Signature:       in.Signature,
		Status:          1,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	logger.Logrus.WithFields(logrus.Fields{"Trade": newTrade}).Info("CreateTradde info")

	//check trade rules
	latestTrade, err := getLatestTrade(newTrade.MinerID, newTrade.TokenAddress)
	if err != nil {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("CreateTrade getLatestTrade failed")
		r.Code = http.StatusInternalServerError
		r.Message = "get latest trade failed"
		return
	}

	errmsg, err := checkTradeValid(latestTrade, newTrade, in.Leverage)
	if err != nil {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("CreateTrade checkTradeValid failed")
		r.Code = http.StatusInternalServerError
		r.Message = errmsg
		return
	}

	err = insertTrade(newTrade)
	if err != nil {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("CreateTrade InsertTrade failed")
		r.Code = http.StatusInternalServerError
		r.Message = "record trade failed"
		return
	}

	err = updatePrice4H(latestTrade, newTrade)
	if err != nil {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Warn("CreateTrade updatePrice4H failed")
	}

	r.Message = "create trade success"
	r.Data = ""
}
