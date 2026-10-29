package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	productv1 "gitlab.ozon.dev/go/classrom/HW2/gen/product/v1"
	"gitlab.ozon.dev/go/classrom/HW3/internal/product"
	"google.golang.org/grpc"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	path := os.Getenv("PRODUCTS_FILE")
	if path == "" {
		path = "products.json"
	}
	products, err := product.LoadFile(path)
	if err != nil {
		return fmt.Errorf("загрузить товары: %w", err)
	}
	address := os.Getenv("PRODUCT_GRPC_ADDR")
	if address == "" {
		address = ":50051"
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("запустить gRPC-сервер: %w", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	log.Printf("Сервис товаров слушает %s; загружено товаров: %d", listener.Addr(), len(products))
	return serve(ctx, listener, products)
}

func serve(ctx context.Context, listener net.Listener, products map[int64]product.Product) error {
	server := grpc.NewServer()
	productv1.RegisterProductServiceServer(server, product.NewServer(products))
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			server.GracefulStop()
		case <-done:
		}
	}()
	if err := server.Serve(listener); err != nil {
		close(done)
		return fmt.Errorf("работа gRPC-сервера: %w", err)
	}
	close(done)
	return nil
}
