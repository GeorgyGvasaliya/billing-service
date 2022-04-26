package db

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"log"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func NewPostgresDB(cfg Config) *sqlx.DB {
	conn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)
	db, err := sqlx.Open("postgres", conn)
	if err != nil {
		log.Println("Cannot connect to db", err)
	}
	log.Println("Connected to db")
	return db
}
