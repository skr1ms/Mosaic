package db

import (
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/skr1ms/mosaic/config"
	"github.com/skr1ms/mosaic/pkg/middleware"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

var logger = middleware.NewLogger()

// Db wraps bun.DB with additional functionality
type Db struct {
	*bun.DB
}

// NewDb creates new database connection with PostgreSQL
func NewDb(config *config.Config) (*Db, error) {
	db, err := pgx.ParseConfig(config.PostgresConfig.URL)
	if err != nil {
		logger.GetZerologLogger().Error().Err(err).Msg("Failed to parse Postgre config")
		return nil, fmt.Errorf("failed to parse Postgres config: %w", err)
	}

	db.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	psqlDB := stdlib.OpenDB(*db)

	logger.GetZerologLogger().Info().Msg("Database connection established successfully")
	return &Db{bun.NewDB(psqlDB, pgdialect.New())}, nil
}
