package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LeUrok/DS-lab2/gateway-service/internal/api"
	"github.com/LeUrok/DS-lab2/gateway-service/internal/client"
	"github.com/LeUrok/DS-lab2/gateway-service/internal/config"
)

func main() {
	cfg := config.Load()

	cars := client.NewCarsClient(cfg.CarsURL)
	rental := client.NewRentalClient(cfg.RentalURL)
	payment := client.NewPaymentClient(cfg.PaymentURL)

	h := api.NewHandler(cars, rental, payment)
	router := api.NewRouter(h)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("gateway-service listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	log.Println("gateway-service stopped")
}
