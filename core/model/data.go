package model

import (
	"time"

	"github.com/uptrace/bun"
)

type AdsTokenEvents struct {
	bun.BaseModel `bun:"table:ads_token_events,alias:oat"`

	TokenAddress string `bun:"token_address,pk,notnull"`
	Chain        string `bun:"chain,pk,notnull"`
	EventID      string `bun:"event_id,pk,notnull"`
	EventType    string `bun:"event_type,pk,notnull"`
	Event        string `bun:"event,pk,notnull"`
	EventDetail  string `bun:"event_detail,notnull"`
	Pt           string `bun:"pt,pk,notnull"`
	BaseScore    string `bun:"base_score,notnull"`
}

type AdsTokenTrade struct {
	bun.BaseModel `bun:"table:ads_token_trades,alias:oat"`

	// ID               int64  `bun:"id,pk,autoincrement"`
	MinerID         string  `bun:"miner_id,pk,notnull"`
	PubKey          string  `bun:"pub_key,notnull"`
	Nonce           int64   `bun:"nonce,pk,notnull"`
	TokenAddress    string  `bun:"token,pk,notnull"`
	PositionManager string  `bun:"position_manager,notnull"`
	Direction       int     `bun:"direction,notnull"`
	Timestamp       int64   `bun:"timestamp,pk,notnull"`
	TradePrice      float64 `bun:"price,notnull"`
	TradePrice4H    float64 `bun:"price_4h"`
	Signature       string  `bun:"signature,notnull"`
	Status          int     `bun:"status,pk,notnull"`
	Leverage        float64 `bun:"leverage"`

	CreatedAt time.Time `bun:"create_at,notnull"`
	UpdatedAt time.Time `bun:"update_at,notnull"`
}

type ResTokenTrade struct {
	bun.BaseModel `bun:"table:ads_token_trades,alias:oat"`

	MinerID         string  `bun:"miner_id,pk,notnull"`
	Nonce           int64   `bun:"nonce,pk,notnull"`
	TokenAddress    string  `bun:"token,pk,notnull"`
	PositionManager string  `bun:"position_manager,notnull"`
	Direction       int     `bun:"direction,notnull"`
	Timestamp       int64   `bun:"timestamp,pk,notnull"`
	TradePrice      float64 `bun:"price,notnull"`
	TradePrice4H    float64 `bun:"price_4h"`
	Leverage        float64 `bun:"leverage"`

	CreatedAt time.Time `bun:"create_at,notnull"`
	UpdatedAt time.Time `bun:"update_at,notnull"`
}

type ChainTokenPrice struct {
	Pt             string  `bun:"pt"`
	Chain          string  `bun:"chain"`
	TokenAddress   string  `bun:"token_address"`
	Price          float64 `bun:"price"`
	Web            string  `bun:"web"`
	ScopeTimeStamp string  `bun:"scope_timestamp"`
	Rank           int     `bun:"rn"`
}

type AdsMinerWhitelist struct {
	bun.BaseModel `bun:"table:ads_addr_whitelist,alias:oat"`

	Address   string `bun:"address,pk,notnull"`
	UID       int    `bun:"uid"`
	Stake     int    `bun:"stake"`
	Status    int    `bun:"status"`
	TimeStamp int64  `bun:"timestamp"`
}

type AdsMinerPerformance struct {
	bun.BaseModel `bun:"table:ads_miner_performance,alias:oat"`

	UID          int    `bun:"uid,pk,notnull"`
	Address      string `bun:"address"`
	RegisterTime string `bun:"register_time"`
}
