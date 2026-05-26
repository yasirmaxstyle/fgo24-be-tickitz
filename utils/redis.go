package utils

import (
	"crypto/tls"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

func InitRedis() *redis.Client {
	addr := fmt.Sprintf("%s:%s",
		Load().RedisHost,
		Load().RedisPort,
	)

	opt := &redis.Options{
		Addr:     addr,
		Password: Load().RedisPassword,
		DB:       0,
	}

	// Enable TLS if the host is not localhost (required for Upstash/Vercel KV)
	if Load().RedisHost != "localhost" && Load().RedisHost != "127.0.0.1" {
		opt.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	client := redis.NewClient(opt)

	log.Println("Redis connected successfully")
	return client
}
