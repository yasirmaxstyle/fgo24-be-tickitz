package main

import (
	"context"
	"log"
	"noir-backend/container"
	"noir-backend/router"
	"noir-backend/seeder"
	"noir-backend/utils"

	"github.com/gin-gonic/gin"
)

//@title NOIR RESTful API
//@version 1.0
//@description backend server of movie ticketing NOIR Project
//@BasePath /

//@securitydefinitions.apikey Token
//@in header
//@name	Authorization

func main() {
	dbpool, err := utils.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer dbpool.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	utils.StartTransactionExpiryJob(ctx, dbpool)
	seeder.SeedTMDBMovies(dbpool)
	seeder.SeedAdminUser(dbpool)

	redis := utils.InitRedis()

	c := container.NewContainer(dbpool, redis)

	r := gin.Default()

	router.CombineRouter(r, c)

	r.Run(":9503")
	log.Println("server runnng on port 9503")
}
