package main

import (
	"flag"
	"log"

	"github.com/Open0xScope/CommuneXService/config"
	"github.com/Open0xScope/CommuneXService/core/task"
	"github.com/Open0xScope/CommuneXService/core/web"
	"github.com/Open0xScope/CommuneXService/utils/logger"
)

func main() {
	configPath := flag.String("config_path", "./", "config file")
	logicLogFile := flag.String("logic_log_file", "./log/communex.log", "logic log file")
	flag.Parse()

	//init logic logger
	logger.Init(*logicLogFile)

	//set log level
	logger.SetLogLevel("debug")

	err := config.LoadConf(*configPath)
	if err != nil {
		log.Fatal("load config failed:", err)
	}

	task.TradeStatusTask()

	task.MinerStatusTask()

	web.Run()
}
