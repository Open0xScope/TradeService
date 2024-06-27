package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Open0xScope/CommuneXService/core/db"
	"github.com/Open0xScope/CommuneXService/core/model"
	"github.com/Open0xScope/CommuneXService/utils/logger"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

var RecordPt string

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	pingPeriod  = 10 * time.Second
	pongTimeout = 60 * time.Second

	subscribers = make(map[*websocket.Conn]bool)
)

func pingLoop(conn *websocket.Conn) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("EventPublish pingLoop failed")
				return
			}
		}
	}
}

type TokenEvents struct {
	TokenAddress string `json:"token_address"`
	Chain        string `json:"chain"`
	EventID      string `json:"event_id"`
	EventType    string `json:"event_type"`
	Event        string `json:"event"`
	EventDetail  string `json:"event_detail"`
	Pt           string `json:"pt"`
	BaseScore    string `json:"base_score"`
}

func getLatestEvents() ([]byte, error) {

	maxpt := ""

	err := db.GetDB().NewSelect().Table("ads_token_events").ColumnExpr("max(pt)").Where("chain in (?)", bun.In(ChainList)).Scan(context.Background(), &maxpt)
	if err != nil {
		return nil, err
	}

	if maxpt <= RecordPt {
		return []byte{}, nil
	}

	res := make([]model.AdsTokenEvents, 0)
	err = db.GetDB().NewSelect().Model(&res).Where("chain in (?) and pt = ? and token_address in (?)", bun.In(ChainList), maxpt, bun.In(TokenList)).Scan(context.Background())
	if err != nil {
		return nil, err
	}

	logger.Logrus.WithFields(logrus.Fields{"Data": res, "MaxPt": RecordPt}).Info("EventPublish getLatestEvents info")

	if len(res) == 0 {
		return []byte{}, nil
	}

	data := make([]TokenEvents, 0)
	for _, v := range res {
		item := TokenEvents{
			TokenAddress: v.TokenAddress,
			Chain:        v.Chain,
			EventID:      v.EventID,
			EventType:    v.EventType,
			Event:        v.Event,
			EventDetail:  v.EventDetail,
			Pt:           v.Pt,
			BaseScore:    v.BaseScore,
		}

		data = append(data, item)
	}

	result, err := json.Marshal(&data)
	if err != nil {
		return nil, err
	}

	RecordPt = maxpt
	return result, nil
}

func EventPublish(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("EventPublish upgrade to WebSocket failed")
		return
	}
	defer conn.Close()

	subscribers[conn] = true
	defer func() {
		delete(subscribers, conn)
	}()

	conn.SetReadDeadline(time.Now().Add(pongTimeout))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongTimeout))
		return nil
	})

	go pingLoop(conn)

	RecordPt = time.Now().UTC().Add(-time.Hour).Format("2006-01-02 15:04:05")[:13]

	// handle message
	for {
		time.Sleep(10 * time.Second)

		// publish event message
		message, err := getLatestEvents()
		if err != nil {
			logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("EventPublish getLatestEvents failed")
			continue
		}

		if len(message) == 0 {
			continue
		}

		for conn := range subscribers {
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err}).Error("EventPublish Failed to write message")
			}
		}
	}
}
