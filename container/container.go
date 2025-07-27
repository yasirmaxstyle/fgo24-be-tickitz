package container

import (
	"noir-backend/controllers"
	"noir-backend/services"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Container struct {
	AuthService           *services.AuthService
	AuthController        *controllers.AuthController
	MovieService          *services.MovieService
	MovieController       *controllers.MovieController
	TransactionService    *services.TransactionService
	TransactionController *controllers.TransactionController
}

func NewContainer(db *pgxpool.Pool, redis *redis.Client) *Container {
	authService := services.NewAuthService(db, redis)
	authController := controllers.NewAuthController(authService)

	movieService := services.NewMovieService(db)
	movieController := controllers.NewMovieController(movieService)

	transactionService := services.NewTransactionService(db)
	transactionController := controllers.NewTransactionController(transactionService)

	return &Container{
		AuthService:           authService,
		AuthController:        authController,
		MovieService:          movieService,
		MovieController:       movieController,
		TransactionService:    transactionService,
		TransactionController: transactionController,
	}
}
