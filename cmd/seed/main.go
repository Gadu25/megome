package main

import (
	"log"
	"megome/config"
	"megome/internal/data/db"
	"megome/internal/data/seeders"

	mysqlCfg "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	config.Load()

	dbConn, err := db.NewMySQLStorage(mysqlCfg.Config{
		User:                 config.Envs.DBUser,
		Passwd:               config.Envs.DBPassword,
		Addr:                 config.Envs.DBAddress,
		DBName:               config.Envs.DBName,
		Net:                  "tcp",
		AllowNativePasswords: true,
		ParseTime:            true,
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := seeders.SeedTechnologies(dbConn); err != nil {
		log.Fatal(err)
	}

	log.Println("Seeding completed.")
}
