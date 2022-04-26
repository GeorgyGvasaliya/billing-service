package main

import (
	"billing-service/db"
	"billing-service/handler"
	"billing-service/utils"
	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
)

func main() {
	if err := utils.InitConfig(); err != nil {
		log.Fatalln("Config init error", err)
	}

	cfg := db.Config{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		User:     viper.GetString("db.username"),
		DBName:   viper.GetString("db.dbname"),
		Password: os.Getenv("DB_PASSWORD"),
	}

	cache := make(map[int]int)

	h := handler.NewHandler(cache, cfg)
	router := httprouter.New()
	h.Register(router)
	log.Fatal(http.ListenAndServe(":"+viper.GetString("server.port"), router))
}
