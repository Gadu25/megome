package main

import (
	"database/sql"
	"fmt"
	"log"
	"megome/cmd/api"
	"megome/config"
	"megome/internal/data/db"

	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	config.Load()

	db, err := db.NewMySQLStorage(mysql.Config{
		User:                 config.Envs.DBUser,
		Passwd:               config.Envs.DBPassword,
		Addr:                 fmt.Sprintf("%s:%s", config.Envs.DBHost, config.Envs.DBPort),
		DBName:               config.Envs.DBName,
		Net:                  "tcp",
		AllowNativePasswords: true,
		ParseTime:            true,
	})

	if err != nil {
		log.Fatal(err)
	}

	initStorage(db)

	server := api.NewAPIServer(":"+config.Envs.Port, db)
	if err := server.Run(); err != nil {
		log.Fatal((err))
	}
}

func initStorage(db *sql.DB) {
	err := db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("DB: Successfully connected!")
}
