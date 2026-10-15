package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	productv1 "gitlab.ozon.dev/go/classrom/HW2/gen/product/v1"
	"gitlab.ozon.dev/go/classrom/HW2/internal/cart"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	productAddress := os.Getenv("PRODUCT_GRPC_ADDR")
	if productAddress == "" {
		productAddress = "localhost:50051"
	}
	conn, err := grpc.NewClient(productAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("создать gRPC-клиент товаров: %w", err)
	}
	defer conn.Close()
	address := os.Getenv("CART_HTTP_ADDR")
	if address == "" {
		address = ":8080"
	}
	server := &http.Server{
		Addr:              address,
		Handler:           cart.Handler(cart.NewService(productv1.NewProductServiceClient(conn)), log.Default()),
		ReadHeaderTimeout: 5 * time.Second,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Не удалось остановить HTTP-сервер: %v", err)
		}
	}()
	log.Printf("Сервис корзины слушает %s; сервис товаров: %s", address, productAddress)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("работа HTTP-сервера: %w", err)
	}
	return nil
}
