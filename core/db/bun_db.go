package db

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/Open0xScope/CommuneXService/config"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

var db *bun.DB
var once sync.Once

var dbPrice *bun.DB
var oncePrice sync.Once

// GetDB get postgressql db instance by sync.Once
func GetDB() *bun.DB {
	once.Do(func() {
		host := config.GetPostgresqlConfig().Host
		port := config.GetPostgresqlConfig().Port
		account := config.GetPostgresqlConfig().Account
		password := config.GetPostgresqlConfig().Password
		dbname := config.GetPostgresqlConfig().DBName
		schemaname := config.GetPostgresqlConfig().SchemaName
		dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&connect_timeout=10", account, password, host, port, dbname)
		sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn), pgdriver.WithConnParams(map[string]interface{}{
			"search_path": schemaname,
		})))

		sqldb.SetMaxOpenConns(10)
		sqldb.SetMaxIdleConns(5)
		sqldb.SetConnMaxLifetime(time.Hour)

		db = bun.NewDB(sqldb, pgdialect.New())
	})
	return db
}
