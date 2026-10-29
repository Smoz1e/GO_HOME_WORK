package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	productv1 "gitlab.ozon.dev/go/classrom/HW2/gen/product/v1"
	"gitlab.ozon.dev/go/classrom/HW3/internal/cart"
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
		productAddress = ":50051"
	}
	target := productAddress
	if strings.HasPrefix(target, ":") {
		target = "localhost" + target
	}
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("создать gRPC-клиент товаров: %w", err)
	}
	defer conn.Close()
	address := os.Getenv("CART_HTTP_ADDR")
	if address == "" {
		address = ":8080"
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("запустить HTTP-сервер: %w", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	log.Printf("Сервис корзины слушает %s; сервис товаров: %s", listener.Addr(), target)
	return serve(ctx, listener, productv1.NewProductServiceClient(conn), log.Default())
}

func serve(ctx context.Context, listener net.Listener, client cart.ProductClient, logger *log.Logger) error {
	server := &http.Server{
		Handler:           cart.Handler(cart.NewService(client), logger),
		ReadHeaderTimeout: 5 * time.Second,
	}
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := server.Shutdown(shutdownCtx); err != nil && logger != nil {
				logger.Printf("Не удалось остановить HTTP-сервер: %v", err)
			}
		case <-done:
		}
	}()
	err := server.Serve(listener)
	close(done)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("работа HTTP-сервера: %w", err)
	}
	return nil
}
