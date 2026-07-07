package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"megome/internal/api"
	"megome/internal/config"
	"megome/internal/db"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	config.Load()

	dbConn, err := db.NewMySQLStorage(mysql.Config{
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

	initStorage(dbConn)

	server := api.NewAPIServer(":"+config.Envs.Port, dbConn)
	if err := server.Run(); err != nil {
		log.Fatal((err))
	}
}

func initStorage(db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := db.PingContext(ctx)
	if err != nil {
		log.Println("WARNING: DB not ready:", err)
	} else {
		log.Println("DB connected")
	}

	log.Println("DB: Successfully connected!")
}
