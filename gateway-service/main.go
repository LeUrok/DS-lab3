package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LeUrok/DS-lab2/gateway-service/internal/api"
	"github.com/LeUrok/DS-lab2/gateway-service/internal/client"
	"github.com/LeUrok/DS-lab2/gateway-service/internal/config"
	"github.com/LeUrok/DS-lab2/gateway-service/internal/queue"
)

func main() {
	cfg := config.Load()

	cars := client.NewCarsClient(cfg.CarsURL)
	rental := client.NewRentalClient(cfg.RentalURL)
	payment := client.NewPaymentClient(cfg.PaymentURL)

	handler := func(ctx context.Context, task queue.Task) error {
		switch task.Type {
		case queue.TaskCancelPayment:
			return payment.Cancel(ctx, task.Payload["paymentUid"])
		case queue.TaskUnreserveCar:
			return cars.Unreserve(ctx, task.Payload["carUid"])
		}
		return fmt.Errorf("unknown task: %s", task.Type)
	}

	q := queue.New(100, handler)

	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	q.Start(workerCtx)

	h := api.NewHandler(cars, rental, payment, q)
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
