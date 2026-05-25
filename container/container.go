package container

import (
	"noir-backend/controllers"
	"noir-backend/repositories"
	"noir-backend/services"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Container struct {
	AuthService           services.AuthService
	AuthController        *controllers.AuthController
	MovieService          services.MovieService
	MovieController       *controllers.MovieController
	TransactionService    services.TransactionService
	TransactionController *controllers.TransactionController
}

func NewContainer(db *pgxpool.Pool, redis *redis.Client) *Container {
	authRepo := repositories.NewAuthRepository(db)
	authService := services.NewAuthService(authRepo, redis)
	authController := controllers.NewAuthController(authService)

	movieRepo := repositories.NewMovieRepository(db)
	movieService := services.NewMovieService(movieRepo)
	movieController := controllers.NewMovieController(movieService)

	transactionRepo := repositories.NewTransactionRepository(db)
	transactionService := services.NewTransactionService(transactionRepo)
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
