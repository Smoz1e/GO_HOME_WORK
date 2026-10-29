package race_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	productv1 "gitlab.ozon.dev/go/classrom/HW2/gen/product/v1"
	"gitlab.ozon.dev/go/classrom/HW3/internal/cart"
	"google.golang.org/grpc"
)

type productClient struct{}

func (productClient) GetProduct(_ context.Context, request *productv1.GetProductRequest, _ ...grpc.CallOption) (*productv1.GetProductResponse, error) {
	return &productv1.GetProductResponse{Sku: request.GetSku(), Name: "Товар", Price: 1}, nil
}

func TestConcurrentAddToCart(t *testing.T) {
	t.Parallel()
	service := cart.NewService(productClient{})
	const workers = 128
	var group sync.WaitGroup
	results := make(chan error, workers)
	for range workers {
		group.Add(1)
		go func() {
			defer group.Done()
			results <- service.Add(context.Background(), 1, 5, 1)
		}()
	}
	group.Wait()
	close(results)
	for err := range results {
		require.NoError(t, err)
	}
	view, err := service.Get(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, view.Items, 1)
	require.Equal(t, uint32(workers), view.Items[0].Count)
	require.Equal(t, uint32(workers), view.TotalPrice)
}
