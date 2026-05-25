package utils

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	DBConnectionString string
	DBSSLMode          string
	RedisHost          string
	RedisPort          string
	RedisPassword      string
	JWTSecret          string
	Port               string
	SMTP               *SMTPConfig
	Admin              *AdminConfig
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type AdminConfig struct {
	Username string
	Email    string
	Password string
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             getEnv("DB_PORT", "5432"),
		DBUser:             getEnv("DB_USER", "postgres"),
		DBPassword:         getEnv("DB_PASSWORD", ""),
		DBName:             getEnv("DB_NAME", "postgres"),
		DBConnectionString: getEnv("DB_CONNECTION_STRING", ""),
		DBSSLMode:          getEnv("DB_SSLMODE", "disable"),
		RedisHost:          getEnv("REDIS_HOST", "localhost"),
		RedisPort:          getEnv("REDIS_PORT", "6379"),
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		JWTSecret:          getEnv("JWT_SECRET", "secretkey"),
		Port:               getEnv("PORT", "8080"),
		SMTP: &SMTPConfig{
			Host:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			Port:     getEnvInt("SMTP_PORT", 587),
			Username: getEnv("SMTP_USERNAME", "yasirmu77@gmail.com"),
			Password: getEnv("SMTP_PASSWORD", "password"),
			From:     getEnv("SMTP_FROM", "yasirmu77@gmail.com"),
		},
		Admin: &AdminConfig{
			Username: getEnv("ADMIN_USERNAME", "admin"),
			Email:    getEnv("ADMIN_EMAIL", "admin@mail.com"),
			Password: getEnv("ADMIN_PASSWORD", "password"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
