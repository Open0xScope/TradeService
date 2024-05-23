package task

import (
	"context"

	"github.com/Open0xScope/CommuneXService/core/db"
	"github.com/Open0xScope/CommuneXService/core/model"
	"github.com/Open0xScope/CommuneXService/utils/logger"
	cron "github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

func MinerStatusTask() {
	c := cron.New()

	_, err := c.AddFunc("@every 10s", func() {
		updateMinerTradeStatus()
	})
	if err != nil {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Fatal("MinerStatusTask start failed")
		return
	}

	c.Start()
}

func updateMinerTradeStatus() error {

	ctx := context.Background()
	var res []model.AdsMinerWhitelist

	err := db.GetDB().NewSelect().Model(&res).Where("status = 0").Scan(ctx)
	if err != nil {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("updateMinerTradeStatus get miner whitelist failed")
		return err
	}

	logger.Logrus.WithFields(logrus.Fields{"NoWhiteList": res}).Info("updateMinerTradeStatus  miner nowhitelist info")

	values := db.GetDB().NewValues(&res)

	rowres, err := db.GetDB().NewUpdate().With("t2", values).Model(&model.AdsTokenTrade{}).TableExpr("t2").Set("status = 0").Where("miner_id = t2.address").Exec(ctx)
	if err != nil {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("updateMinerTradeStatus get miner whitelist failed")
		return err
	}
	logger.Logrus.WithFields(logrus.Fields{"UpdateStatusResult": rowres}).Info("updateMinerTradeStatus update trade status result")

	return nil
}
