package main

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	productv1 "gitlab.ozon.dev/go/classrom/HW2/gen/product/v1"
	"gitlab.ozon.dev/go/classrom/HW3/internal/product"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func TestServeProduct(t *testing.T) {
	t.Parallel()
	listener := bufconn.Listen(1024 * 1024)
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		result <- serve(ctx, listener, map[int64]product.Product{5: {SKU: 5, Name: "Пять", Price: 50}})
	}()
	t.Cleanup(func() {
		cancel()
		_ = listener.Close()
		<-result
	})
	conn, err := grpc.NewClient("passthrough:///product", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })
	requestCtx, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()
	response, err := productv1.NewProductServiceClient(conn).GetProduct(requestCtx, &productv1.GetProductRequest{Sku: 5})
	require.NoError(t, err)
	require.Equal(t, "Пять", response.GetName())
}
