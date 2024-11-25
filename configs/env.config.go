package configs

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type MySQLConfig struct {
	DBUser     string
	DBPassword string
	DBName     string
	DBHost     string
	DBPort     string
}

type CassandraConfig struct {
}

func LoadMySQLConfig() *MySQLConfig {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	config := &MySQLConfig{
		DBUser:     os.Getenv("MYSQL_DB_USER"),
		DBPassword: os.Getenv("MYSQL_DB_PASSWORD"),
		DBName:     os.Getenv("MYSQL_DB_NAME"),
		DBHost:     os.Getenv("MYSQL_DB_HOST"),
		DBPort:     os.Getenv("MYSQL_DB_PORT"),
	}
	return config
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
