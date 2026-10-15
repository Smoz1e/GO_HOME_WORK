package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	productv1 "gitlab.ozon.dev/go/classrom/HW2/gen/product/v1"
	"gitlab.ozon.dev/go/classrom/HW2/internal/product"
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
	server := grpc.NewServer()
	productv1.RegisterProductServiceServer(server, product.NewServer(products))
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)
	go func() { <-stop; server.GracefulStop() }()
	log.Printf("Сервис товаров слушает %s; загружено товаров: %d", listener.Addr(), len(products))
	if err := server.Serve(listener); err != nil {
		return fmt.Errorf("работа gRPC-сервера: %w", err)
	}
	return nil
}
