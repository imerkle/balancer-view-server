package db

import (
	"balancer-view/utils"
	"context"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
)

var Dbpool *pgxpool.Pool
var DefaultPostgresConn = "postgres://postgres:postgres@localhost:5433/postgres?pool_max_conns=20"
var err error

func Connect() {
	connStr := utils.GetEnv("POSTGRES_CONN", DefaultPostgresConn)
	Dbpool, err = pgxpool.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatal("Unable to connect to db ", connStr)
	}
}
