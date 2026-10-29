package main

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	productv1 "gitlab.ozon.dev/go/classrom/HW2/gen/product/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type fakeProductClient struct{}

func (fakeProductClient) GetProduct(_ context.Context, request *productv1.GetProductRequest, _ ...grpc.CallOption) (*productv1.GetProductResponse, error) {
	return &productv1.GetProductResponse{Sku: request.GetSku(), Name: "Товар", Price: 10}, nil
}

func TestServeCart(t *testing.T) {
	t.Parallel()
	listener := bufconn.Listen(1024 * 1024)
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		result <- serve(ctx, listener, fakeProductClient{}, log.New(io.Discard, "", 0))
	}()
	t.Cleanup(func() {
		cancel()
		_ = listener.Close()
		<-result
	})
	client := &http.Client{Transport: &http.Transport{DialContext: func(context.Context, string, string) (net.Conn, error) {
		return listener.Dial()
	}}}
	t.Cleanup(client.CloseIdleConnections)
	requestCtx, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()
	request, err := http.NewRequestWithContext(requestCtx, http.MethodGet, "http://cart/user/1/cart", nil)
	require.NoError(t, err)
	response, err := client.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, http.StatusNotFound, response.StatusCode)
}
