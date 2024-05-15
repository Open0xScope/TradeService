package task

import (
	"context"
	"fmt"
	"time"

	"github.com/Open0xScope/CommuneXService/core/db"
	"github.com/Open0xScope/CommuneXService/core/model"
	"github.com/Open0xScope/CommuneXService/utils/logger"
	cron "github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

func TradeStatusTask() {
	c := cron.New()

	_, err := c.AddFunc("@every 10s", func() {
		updateTrade()
	})
	if err != nil {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Fatal("TradeStatusTask start failed")
		return
	}

	c.Start()
}

func updateTrade() error {
	ctx := context.Background()
	var res []model.AdsTokenTrade

	after4h := time.Now().Add(-4 * time.Hour).Unix()

	err := db.GetDB().NewSelect().Model(&res).Where("price_4h is null or price_4h = 0 and timestamp <= ?", after4h).Order("timestamp asc").Scan(ctx)
	if err != nil {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("TradeStatusTask get trades failed")
		return err
	}

	for _, order := range res {
		err = updateTradePrice4h(order)
		if err != nil {
			logger.Logrus.WithFields(logrus.Fields{"Trade": order, "WarnMsg": err}).Warn("TradeStatusTask warn")
			continue
		}

		logger.Logrus.WithFields(logrus.Fields{"Trade": order}).Info("TradeStatusTask info")
	}

	return nil
}

func updateTradePrice4h(order model.AdsTokenTrade) error {
	stamp := order.Timestamp + int64(14400)
	priceObj, err := getTokenPrice(order.TokenAddress, stamp)
	if err != nil {
		return fmt.Errorf("get token price,%v", err)
	}

	order.TradePrice4H = priceObj.Price

	_, err = db.GetDB().NewUpdate().Model(&order).Set("price_4h = ?", priceObj.Price).Where("miner_id = ? and token = ? and nonce = ?", order.MinerID, order.TokenAddress, order.Nonce).Exec(context.Background())
	if err != nil {
		return fmt.Errorf("update trade 4h price,%v", err)
	}

	return nil
}

func getTokenPrice(token string, timestamp int64) (*model.ChainTokenPrice, error) {
	ctx := context.Background()
	var res model.ChainTokenPrice

	mathtime := time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")

	err := db.GetDB().NewSelect().Model(&res).Where("chain = ? and token_address = ? and pt <= ?", "eth", token, mathtime).Order("pt DESC").Limit(1).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
