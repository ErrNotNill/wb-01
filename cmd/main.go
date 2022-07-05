package main

import (
	"context"
	"github.com/nats-io/nats.go"
	"github.com/patrickmn/go-cache"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"wb-l0/config"
	"wb-l0/logger"
	"wb-l0/pkg/handler"
	"wb-l0/pkg/repository"
	"wb-l0/server"
)

func main() {
	log := logger.InitLogger()
	defer func() {
		if err := log.Sync(); err != nil {
			zap.L().Error("error syncing logger", zap.Error(err))
		}
	}()

	if err := config.InitConfig(); err != nil {
		zap.L().Error("error initializing config: %s", zap.Error(err))
	}

	zap.L().Info("config initialized")

	natsConn, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		zap.L().Error("error connecting nats", zap.Error(err))
	}

	nc, err := nats.NewEncodedConn(natsConn, nats.JSON_ENCODER)
	if err != nil {
		zap.L().Error("error connecting encoding nats", zap.Error(err))
	}

	zap.L().Info("nats initialized")

	c := cache.New(5*time.Minute, 10*time.Minute)

	zap.L().Info("cache initialized")

	db, err := repository.NewPostgresDb(viper.GetString("db.uri"))
	if err != nil {
		zap.L().Fatal("failed to initializing database: %s", zap.Error(err))
	}

	pdb := repository.NewRepository(db)

	handlers := handler.NewHandler(pdb, c, nc)

	srv := new(server.Server)

	go func() {
		if err := srv.Run(viper.GetString("server.port"), handlers.InitRoutes()); err != nil &&
			err != http.ErrServerClosed {
			zap.L().Error("error running server: %s", zap.Error(err))
		}
	}()

	zap.L().Info("server started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	if err := srv.Shutdown(context.Background()); err != nil {
		zap.L().Error("error shutting down server: %s", zap.Error(err))
	}

	zap.L().Info("server shut down")

	if err = db.Close(context.Background()); err != nil {
		zap.L().Fatal("error close database", zap.Error(err))
	}

}
