package api

import (
	"fmt"
	"log"
	"net/http"
	"noir-backend/container"
	"noir-backend/router"
	"noir-backend/utils"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	app      *gin.Engine
	initOnce sync.Once
	initErr  error
)

func initializeApp() {
	// 1. Connect to Database
	dbpool, err := utils.ConnectDB()
	if err != nil {
		initErr = fmt.Errorf("failed to connect to DB: %v", err)
		return
	}

	// 2. Initialize Redis
	redis := utils.InitRedis()

	// 3. Create Container
	c := container.NewContainer(dbpool, redis)

	// 4. Setup Gin Engine
	app = gin.Default()

	// 5. Apply Routes
	router.CombineRouter(app, c)
}

// Handler is the Vercel serverless function entrypoint
func Handler(w http.ResponseWriter, r *http.Request) {
	initOnce.Do(initializeApp)

	if initErr != nil {
		log.Printf("Initialization error: %v", initErr)
		http.Error(w, "Internal Server Error: Application failed to initialize", http.StatusInternalServerError)
		return
	}

	app.ServeHTTP(w, r)
}
