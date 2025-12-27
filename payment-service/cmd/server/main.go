package main

import (
	"common_library/logging"
	"context"
	"github.com/niklvrr/OnlineShop/payment-service/internal/config"
	"github.com/niklvrr/OnlineShop/payment-service/internal/handler"
	"github.com/niklvrr/OnlineShop/payment-service/internal/infrastructure/kafka"
	"github.com/niklvrr/OnlineShop/payment-service/internal/infrastructure/pgdb"
	"github.com/niklvrr/OnlineShop/payment-service/internal/usecase"
	paymentpb "github.com/niklvrr/OnlineShop/proto-contracts/gen/go/payment"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
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

	paymentRepo := pgdb.NewPaymentRepository(pgClient)
	paymentUC := usecase.NewPaymentUseCase(paymentRepo, appLogger)
	paymentHandler := handler.NewPaymentHandler(paymentUC, appLogger)

	grpcServer := grpc.NewServer()
	paymentpb.RegisterPaymentServiceServer(grpcServer, paymentHandler)

	lis, err := net.Listen("tcp", ":"+cfg.Grpc.Port)
	if err != nil {
		appLogger.Error("failed to listen", err)
		return
	}

	kafkaConsumer, err := kafka.NewConsumer(cfg.Kafka.Brokers, "payment-service-group", paymentUC, appLogger)
	if err != nil {
		appLogger.Error("failed to create kafka consumer", err)
		return
	}

	go func() {
		topics := []string{"order-events"}
		if err := kafkaConsumer.Start(ctx, topics); err != nil {
			appLogger.Error("kafka consumer error", err)
		}
	}()

	go func() {
		appLogger.Info("gRPC server starting", "port", cfg.Grpc.Port)
		if err := grpcServer.Serve(lis); err != nil {
			appLogger.Error("failed to serve", err)
			return
		}
	}()

	appLogger.Info("payment service started")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	appLogger.Info("shutting down payment service")
	grpcServer.GracefulStop()
}

