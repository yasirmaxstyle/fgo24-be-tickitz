package main

import (
	"log"
	"noir-backend/container"
	"noir-backend/router"
	"noir-backend/seeder"
	"noir-backend/utils"
	"os"

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

	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	// Parse flags for seeding
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "--seed" {
		log.Println("Running seeders...")
		err := seeder.SeedTMDBMovies(dbpool)
		if err != nil {
			log.Printf("Error seeding movies: %v\n", err)
		}
		seeder.SeedAdminUser(dbpool)
		log.Println("Seeding complete. Exiting...")
		return
	}

	redis := utils.InitRedis()

	c := container.NewContainer(dbpool, redis)

	r := gin.Default()

	router.CombineRouter(r, c)

	log.Println("Server running on http://localhost:9503")
	log.Println("Swagger documentation available at http://localhost:9503/swagger/index.html")
	r.Run(":9503")
}
