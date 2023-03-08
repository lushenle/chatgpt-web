package main

import (
	"database/sql"
	"flag"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/lushenle/chatgpt-web/api"
	db "github.com/lushenle/chatgpt-web/db/sqlc"
	"github.com/lushenle/chatgpt-web/util"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "config", "Configuration file path")
	flag.Parse()
}

func main() {
	gin.SetMode(gin.DebugMode)

	config, err := util.LoadConfig(configPath)
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	conn, err := sql.Open(config.ChatGPT.DBDriver, config.ChatGPT.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	err = server.Start(config.ChatGPT.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
