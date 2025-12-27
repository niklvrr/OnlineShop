package main

import (
	"common_library/logging"
	"github.com/gorilla/mux"
	"github.com/niklvrr/OnlineShop/api-gateway/internal/client"
	"github.com/niklvrr/OnlineShop/api-gateway/internal/config"
	"github.com/niklvrr/OnlineShop/api-gateway/internal/handler"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.LoadConfig()
	appLogger := logging.NewLogger("debug")

	orderClient, err := client.NewOrderClient(cfg.OrderServiceURL)
	if err != nil {
		log.Fatalf("failed to create order client: %v", err)
	}
	defer orderClient.Close()

	paymentClient, err := client.NewPaymentClient(cfg.PaymentServiceURL)
	if err != nil {
		log.Fatalf("failed to create payment client: %v", err)
	}
	defer paymentClient.Close()

	orderHandler := handler.NewOrderHandler(orderClient, appLogger)
	paymentHandler := handler.NewPaymentHandler(paymentClient, appLogger)

	router := mux.NewRouter()
	api := router.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/orders", orderHandler.CreateOrder).Methods("POST")
	api.HandleFunc("/orders/{order_id}", orderHandler.GetOrderStatus).Methods("GET")

	api.HandleFunc("/payments", paymentHandler.InitiatePayment).Methods("POST")
	api.HandleFunc("/payments/{payment_id}", paymentHandler.GetPaymentStatus).Methods("GET")

	srv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: router,
	}

	go func() {
		appLogger.Info("API Gateway starting", "port", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Error("failed to start server", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	appLogger.Info("shutting down API Gateway")
}

