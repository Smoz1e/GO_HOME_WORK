package api

import (
	"os"

	productv1 "gitlab.ozon.dev/go/classrom/HW2/gen/product/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func connectProduct() (productv1.ProductServiceClient, *grpc.ClientConn, error) {
	address := os.Getenv("PRODUCT_GRPC_ADDR")
	if address == "" {
		address = "localhost:50051"
	}
	if address[0] == ':' {
		address = "localhost" + address
	}
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return productv1.NewProductServiceClient(conn), conn, nil
}

func cartURL() string {
	address := os.Getenv("CART_HTTP_URL")
	if address == "" {
		return "http://localhost:8080"
	}
	return address
}
