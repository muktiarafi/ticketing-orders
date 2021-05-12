package config

import (
	"fmt"
	"os"
)

func PostgresDSN() string {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")

	return fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s",
		host,
		port,
		dbName,
		user,
		password,
	)
}
