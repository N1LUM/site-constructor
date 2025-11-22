package app

import (
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"site-constructor/configs"
	"site-constructor/internal/auth"
	"site-constructor/internal/database"
	"site-constructor/internal/handler"
	"site-constructor/internal/repository"
	"site-constructor/internal/service"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

func Run() {
	logrus.SetFormatter(new(logrus.JSONFormatter))

	redisConfig := configs.InitRedisConfig()

	redis, err := database.ConnectRedis(redisConfig)
	if err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}

	postgresConfig := configs.InitPostgresConfig()

	postgres, err := database.ConnectPostgres(postgresConfig)
	if err != nil {
		log.Fatalf("Failed to connect to postgres: %v", err)
	}

	database.MigratePostgres(postgres)

	jwtManager := auth.NewJWTManager(
		[]byte(os.Getenv("JWT_ACCESS_SECRET")),
		[]byte(os.Getenv("JWT_REFRESH_SECRET")),
	)

	repositories := repository.NewRepository(postgres, redis)
	services := service.NewService(repositories, jwtManager)
	handlers := handler.NewHandler(services, jwtManager)

	server := new(configs.Server)
	address := configs.BuildAppAddress()

	go func() {
		logrus.Infof("Starting server on %s", address)

		if err := server.Run(address, handlers.InitRoutes()); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.Fatalf("Error running HTTP server: %s", err)
		}
	}()

	gracefulShutdown(server)
}

func gracefulShutdown(server *configs.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logrus.Fatalf("Server forced to shutdown: %s", err)
	}

	logrus.Info("Server exited gracefully")
}
