package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/http"
	"runtime"
	"time"

	"github.com/Open0xScope/CommuneXService/core/db"
	"github.com/Open0xScope/CommuneXService/core/model"
	"github.com/Open0xScope/CommuneXService/core/redis"
	"github.com/Open0xScope/CommuneXService/utils/logger"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type InCreateTrade struct {
	MinerID         string  `json:"miner_id"`
	PubKey          string  `json:"pub_key"`
	Nonce           int64   `json:"nonce"`
	Token           string  `json:"token"`
	PositionManager string  `json:"position_manager"`
	Direction       int     `json:"direction"`
	Timestamp       int64   `json:"timestamp"`
	Leverage        float64 `json:"leverage"`
	Signature       string  `json:"signature"`
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

func checknewtrade(newTrade *model.AdsTokenTrade) (string, error) {
	_, err := IsMinerOrValidor(newTrade.MinerID)
	if err != nil {
		return "miner not registered", err
	}

	err = CheckAddress(newTrade.PubKey, newTrade.MinerID)
	if err != nil {
		return "address and key not match", err
	}

	msg := fmt.Sprintf("%s%s%d%s%s%d%d", newTrade.MinerID, newTrade.PubKey, newTrade.Nonce, newTrade.TokenAddress, newTrade.PositionManager, newTrade.Direction, newTrade.Timestamp)
	if newTrade.Leverage != 0 {
		msg += fmt.Sprintf("%v", newTrade.Leverage)
	}

	err = VerifySign(msg, newTrade.PubKey, newTrade.Signature)
	if err != nil {
		return "sign error", err
	}

	err = CheckTradeRateLimit(newTrade.PubKey)
	if err != nil {
		return "access limit exceeded, please try again later", err
	}

	err = CheckTradeRateLimitDay(newTrade.PubKey)
	if err != nil {
		return "user access day limit exceeded, please try again later", err
	}

	err = CheckTradeTokenRateLimitDay(newTrade.PubKey, newTrade.TokenAddress)
	if err != nil {
		return "token access day limit exceeded, please try again later", err
	}

	leverage := newTrade.Leverage
	if leverage != 0 {
		if newTrade.TokenAddress == "0x0000000000000000000000000000000000000000" || newTrade.TokenAddress == "0x2260fac5e5542a773aa44fbcfedf7c193bc2c599" {
			if leverage < 0.1 || leverage > 50 {
				return "the leverage is not in the range for eth/btc", errors.New("the leverage is not in the range")
			}
		} else {
			if leverage < 0.1 || leverage > 20 {
				return "the leverage is not in the range for others", errors.New("the leverage is not in the range")
			}
		}

		if !isDivisible(leverage, 0.1) {
			return "the leverage is not an integer multiple of the basic unit", errors.New("the leverage is not an integer multiple of the basic unit")
		}
	} else {
		leverage = float64(1.0)
	}

	newTrade.Leverage = leverage

	if newTrade.PositionManager != "open" && newTrade.PositionManager != "close" {
		return "position manager invalid", errors.New("position manager is invalid")
	}

	if newTrade.Timestamp > time.Now().Unix() {
		return "creation timestamp more than now", errors.New("creation timestamp more than now")
	}

	last10min := time.Now().Add(-time.Minute).Unix()
	if newTrade.Timestamp < last10min {
		return "creation timestamp is old", errors.New("creation timestamp is old")
	}

	return "check new trade success", nil
}

func checktwotrade(latestTrade, newTrade *model.AdsTokenTrade) (string, error) {
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

	return "check two trade success", nil
}

func checkTradeLeverageLimit(oldTrade, newTrade *model.AdsTokenTrade) (string, error) {
	rkey := fmt.Sprintf("%s%s%d", newTrade.MinerID, newTrade.TokenAddress, newTrade.Direction)

	cv, err := redis.GetCounterValue(rkey)
	if err != nil {
		return "get key value failed", err
	}
	if cv >= 7 {
		err = redis.SetCounterExpir(rkey, 7*24*time.Hour)
		if err != nil {
			return "open trade limit exceeded and set time failed", err
		}
		return "open trade limit exceeded", errors.New("open trade limit exceeded")
	}

	if oldTrade.Leverage >= newTrade.Leverage {
		err = redis.DelCounter(rkey)
		if err != nil {
			return "del key failed", err
		}
	} else {
		err = redis.SetCounter(rkey, cv+1)
		if err != nil {
			return "set key value failed", err
		}
	}

	return insertCloseTrade(oldTrade)
}

func insertCloseTrade(trade *model.AdsTokenTrade) (string, error) {
	//insert close trade
	closeTrade := &model.AdsTokenTrade{
		MinerID:         trade.MinerID,
		PubKey:          trade.PubKey,
		Nonce:           trade.Nonce + 1,
		TokenAddress:    trade.TokenAddress,
		PositionManager: "close",
		Direction:       trade.Direction,
		Timestamp:       trade.Timestamp + 1,
		TradePrice:      trade.TradePrice,
		Signature:       "no need sign",
		Status:          1,
		Leverage:        trade.Leverage,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	err := insertTrade(closeTrade)
	if err != nil {
		return "insert close trade failed", err
	}

	return "insert close trade success", nil
}

func checkTrade(oldtrade, newTrade *model.AdsTokenTrade) (string, error) {
	msg, err := checknewtrade(newTrade)
	if err != nil {
		return msg, err
	}

	if newTrade.PositionManager == "open" && oldtrade != nil && newTrade.PositionManager == oldtrade.PositionManager {
		return checkTradeLeverageLimit(oldtrade, newTrade)
	}

	return checktwotrade(oldtrade, newTrade)
}

func updatePrice4H(latestTrade, newTrade *model.AdsTokenTrade) error {
	if latestTrade == nil {
		return nil
	}

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

func PrintStack() string {
	var buf [4096]byte
	n := runtime.Stack(buf[:], false)
	return string(buf[:n])
}

func CreateTradde(c *gin.Context) {
	r := &Response{
		Code:    http.StatusOK,
		Message: "success",
	}
	defer func(r *Response) {
		err := recover()
		if err != nil {
			logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err, "Stack": PrintStack()}).Fatalf("TradeStatusTask panic")
			c.JSON(http.StatusInternalServerError, r)
		} else {
			c.JSON(http.StatusOK, r)
		}
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
		Leverage:        in.Leverage,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	logger.Logrus.WithFields(logrus.Fields{"Trade": newTrade}).Info("CreateTradde info")

	//check trade rules
	oldTrade, err := getLatestTrade(newTrade.MinerID, newTrade.TokenAddress)
	if err != nil {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("CreateTrade getLatestTrade failed")
		r.Code = http.StatusInternalServerError
		r.Message = "get latest trade failed"
		return
	}

	errmsg, err := checkTrade(oldTrade, newTrade)
	if err != nil {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("CreateTrade checkTrade failed")
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

	err = updatePrice4H(oldTrade, newTrade)
	if err != nil {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Warn("CreateTrade updatePrice4H failed")
	}

	r.Message = "create trade success"
	r.Data = ""
}
