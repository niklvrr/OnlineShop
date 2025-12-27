package main

import (
	"common_library/logging"
	"context"
	"github.com/niklvrr/OnlineShop/order-service/internal/config"
	"github.com/niklvrr/OnlineShop/order-service/internal/handler"
	"github.com/niklvrr/OnlineShop/order-service/internal/infrastructure/kafka"
	"github.com/niklvrr/OnlineShop/order-service/internal/infrastructure/pgdb"
	"github.com/niklvrr/OnlineShop/order-service/internal/usecase"
	orderpb "github.com/niklvrr/OnlineShop/proto-contracts/gen/go/order"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	ctx := context.Background()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("config init error %v", err)
		return
	}

	appLogger := logging.NewLogger(cfg.App.EnvLevel)

	pgClient, err := pgdb.NewPgClient(cfg.Database.URL)
	if err != nil {
		appLogger.Error(err.Error(), err)
		return
	}
	defer pgClient.Close()
	appLogger.Debug("db init success")

	err = pgdb.RunMigrations(cfg.Database.URL, appLogger)
	if err != nil {
		appLogger.Error(err.Error(), err)
		return
	}
	appLogger.Debug("migration run success")

	orderRepo := pgdb.NewOrderRepository(pgClient)
	orderUC := usecase.NewOrderUseCase(orderRepo, appLogger)
	orderHandler := handler.NewOrderHandler(orderUC, appLogger)

	grpcServer := grpc.NewServer()
	orderpb.RegisterOrderServiceServer(grpcServer, orderHandler)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		appLogger.Error("failed to listen", err)
		return
	}

	brokers := strings.Split(strings.Join(cfg.Kafka.AdvertisedListeners, ","), ",")
	dispatcher, err := kafka.NewDispatcher(brokers, orderRepo, appLogger, "order-events")
	if err != nil {
		appLogger.Error("failed to create kafka dispatcher", err)
		return
	}
	defer dispatcher.Close()

	go dispatcher.Start(ctx)

	go func() {
		appLogger.Info("gRPC server starting", "port", "50051")
		if err := grpcServer.Serve(lis); err != nil {
			appLogger.Error("failed to serve", err)
			return
		}
	}()

	appLogger.Info("order service started")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	appLogger.Info("shutting down order service")
	grpcServer.GracefulStop()
}
