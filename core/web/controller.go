package web

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Open0xScope/CommuneXService/core/web/handler"
	"github.com/Open0xScope/CommuneXService/utils/logger"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func ServerRoute() *gin.Engine {
	router := gin.New()

	recoverFile, err := os.OpenFile("./log/recover.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil || recoverFile == nil {

		if err != nil {
			logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err.Error()}).Error("open recover log file failed")
		}
		if recoverFile == nil {
			logger.Logrus.Error("open recover log file failed:recoverFile is nil")
		}

		return nil
	}

	router.Use(MiddleLogger("./log/visit.log"), gin.RecoveryWithWriter(recoverFile))

	// http router
	router.POST("/createtrade", handler.CreateTradde)
	router.GET("/getusertrades", handler.GetUserTraddes)
	router.GET("/getalltrades", handler.GetAllTraddes)
	router.GET("/getregistertime", handler.GetRegisterTime)

	router.GET("/getallevents", handler.GetAllEvents)
	router.GET("/getlatestprice", handler.GetLatestPrice)

	// WebSocket 路由
	router.GET("/ws/getevents", handler.EventPublish)

	return router
}

func Run() {
	router := ServerRoute()
	if router != nil {
		server := &http.Server{
			Addr:         ":8000",
			Handler:      router,
			ReadTimeout:  120 * time.Second,
			WriteTimeout: 120 * time.Second,
		}

		go func() {
			err := server.ListenAndServe()
			if err != nil {
				logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Fatal("Server start failed")
			}
		}()

		// Wait for interrupt signal to gracefully shutdown the server with
		// a timeout of 5 seconds.
		quit := make(chan os.Signal)
		// kill (no param) default send syscall.SIGTERM
		// kill -2 is syscall.SIGINT
		// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		// The context is used to inform the server it has 5 seconds to finish
		// the request it is currently handling
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err.Error()}).Error("Server forced to shutdown")
		}

		logger.Logrus.Info("Server start success")
	}
}
