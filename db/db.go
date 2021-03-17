package db

import (
	"balancer-view/utils"
	"context"
	"log"

	"github.com/hasura/go-graphql-client"
	"github.com/jackc/pgx/v4/pgxpool"
)

var Dbpool *pgxpool.Pool
var defaultPostgresConn = "postgres://postgres:postgres@localhost:5433/postgres"
var err error
var Gqlclient *graphql.Client

func Connect() {
	connStr := utils.GetEnv("POSTGRES_CONN", defaultPostgresConn)
	Dbpool, err = pgxpool.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatal("Unable to connect to db ", connStr)
	}
}

func InitGQL() {
	Gqlclient = graphql.NewClient("https://api.thegraph.com/subgraphs/name/balancer-labs/balancer", nil)
}
