package db

import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/skr1ms/mosaic/config"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

type Db struct {
	*bun.DB
}

func NewDb(config *config.Config) *Db {
	db, err := pgx.ParseConfig(config.DatabaseConfig.URL)
	if err != nil {
		panic(err)
	}
	db.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	psqlDB := stdlib.OpenDB(*db)
	return &Db{bun.NewDB(psqlDB, pgdialect.New())}
}
