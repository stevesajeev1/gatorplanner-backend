package dependencies

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stevesajeev1/gatorplanner-backend/internal/database/sqlc"
)

type DB struct {
	Pool  *pgxpool.Pool
	Query *sqlc.Queries
}

func NewDB(connStr string) *DB {
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Panic(err)
	}

	query := sqlc.New(pool)

	db := DB{
		Pool:  pool,
		Query: query,
	}

	return &db
}

func (d *DB) Close() {
	d.Pool.Close()
}
