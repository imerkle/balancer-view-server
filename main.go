package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"balancer-view/config"
	"balancer-view/db"
	"balancer-view/syncer"
	"balancer-view/utils"

	"log"
	"net/http"
)

var workspaceRoot = utils.GetEnv("WORKSPACE_ROOT", "")

func main() {

	db.Connect()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Dbpool.Close()

	var yconf config.YamlConfig
	yconf.GetConf(workspaceRoot)
	config.ConfInit(yconf)

	yconf.SyncInterval = utils.GetEnvAsInt("SYNC_INTERVAL", yconf.SyncInterval)
	yconf.BatchDays = utils.GetEnvAsInt("BATCH_DAYS", yconf.BatchDays)
	yconf.ResetDb = utils.GetEnvAsBool("RESET_DB", yconf.ResetDb)

	if yconf.ResetDb {
		queryDeleteTable := `DROP TABLE swaps; DROP TABLE pairs;`
		_, err = db.Dbpool.Exec(context.Background(), queryDeleteTable)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to delete tables: %v\n", err)
		} else {
			fmt.Println("Successfully deleted tables")
		}
	}
	queryCreateTable := `CREATE TABLE swaps (id VARCHAR(150), pair VARCHAR(50), price FLOAT, amount FLOAT, time TIMESTAMPTZ NOT NULL); SELECT create_hypertable('swaps', 'time');CREATE TABLE pairs (pair VARCHAR(50));`
	_, err = db.Dbpool.Exec(context.Background(), queryCreateTable)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create SWAPS table: %v\n", err)
	} else {
		fmt.Println("Successfully created relational table SWAPS")
	}

	syncerGroup := &syncer.SyncerGroup{SyncInterval: yconf.SyncInterval}
	go syncerGroup.Init(yconf.BatchDays)

	// Register http handlers
	registerHanders(map[string]func(http.ResponseWriter, *http.Request){
		"/": func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, `Datafeed version is 2.0.4
			Valid keys count is 5
			Current key is zy1`, ok200)
		},
		"/config": func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, config.Conf, ok200)
		},
		"/symbol_info": respondNotImplemented,
		"/symbols":     symbolsHandler,
		"/search":      searchHandler,
		"/history":     historyHandler,
		"/time": func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, time.Now().UTC().Unix(), ok200)
		},
	})

	args := os.Args[1:]
	port := utils.GetEnv("PORT", "80")
	if len(args) > 0 {
		port = args[0]
	}
	log.Println("Balancer View data feed server started at port ", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

}
