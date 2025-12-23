package main

import (
	"common_library/logging"
	"context"
	"github.com/niklvrr/OnlineShop/order-service/internal/config"
	"github.com/niklvrr/OnlineShop/order-service/internal/infrastructure/pgdb"
	"log"
)

func main() {
	// TODO context init
	ctx := context.Background()

	// TODO config init
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("config init error %v", err)
		return
	}

	// TODO logger init
	appLogger := logging.NewLogger(cfg.App.EnvLevel)

	// TODO database init
	pgClient, err := pgdb.NewPgClient(cfg.Database.URL)
	if err != nil {
		appLogger.Error(err.Error(), err)
	}
	defer pgClient.Close()
	appLogger.Debug("db init success")

	err = pgdb.RunMigrations(cfg.Database.URL, appLogger)
	if err != nil {
		appLogger.Error(err.Error(), err)
	}
	appLogger.Debug("migration run success")

	// TODO kafka client init

	// TODO payment service client init

	// TODO all layers init

	// TODO gRPC server init
}
