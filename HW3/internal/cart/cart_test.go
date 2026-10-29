package cart_test

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	productv1 "gitlab.ozon.dev/go/classrom/HW2/gen/product/v1"
	"gitlab.ozon.dev/go/classrom/HW3/internal/cart"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeClient struct {
	mu        sync.RWMutex
	products  map[int64]*productv1.GetProductResponse
	deadlines []time.Time
	delay     time.Duration
	calls     atomic.Int32
	active    atomic.Int32
	maximum   atomic.Int32
}

func newFake() *fakeClient {
	return &fakeClient{products: map[int64]*productv1.GetProductResponse{
		2: {Sku: 2, Name: "Два", Price: 20},
		5: {Sku: 5, Name: "Пять", Price: 50},
		9: {Sku: 9, Name: "Девять", Price: 90},
	}}
}

func (f *fakeClient) GetProduct(ctx context.Context, request *productv1.GetProductRequest, _ ...grpc.CallOption) (*productv1.GetProductResponse, error) {
	f.calls.Add(1)
	active := f.active.Add(1)
	for maximum := f.maximum.Load(); active > maximum; maximum = f.maximum.Load() {
		if f.maximum.CompareAndSwap(maximum, active) {
			break
		}
	}
	defer f.active.Add(-1)
	deadline, ok := ctx.Deadline()
	if ok {
		f.mu.Lock()
		f.deadlines = append(f.deadlines, deadline)
		f.mu.Unlock()
	}
	if f.delay > 0 {
		select {
		case <-time.After(f.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	f.mu.RLock()
	product := f.products[request.GetSku()]
	f.mu.RUnlock()
	if product == nil {
		return nil, status.Error(codes.NotFound, "товар не найден")
	}
	return product, nil
}

func TestAddToCart(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		userID int64
		sku    int64
		count  uint32
		want   error
	}{
		{"нулевой пользователь", 0, 2, 1, cart.ErrInvalidUser},
		{"нулевой SKU", 1, 0, 1, cart.ErrInvalidSKU},
		{"нулевое количество", 1, 2, 0, cart.ErrInvalidCount},
		{"нет товара", 1, 7, 1, cart.ErrProductNotFound},
		{"существующий товар", 1, 2, 1, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			service := cart.NewService(newFake())
			err := service.Add(context.Background(), tc.userID, tc.sku, tc.count)
			if tc.want == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.want)
			}
		})
	}
}

func TestGetCartCases(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name     string
		userID   int64
		populate bool
		remove   bool
		want     error
	}{
		{"нулевой пользователь", 0, false, false, cart.ErrInvalidUser},
		{"пустая корзина", 1, false, false, cart.ErrCartNotFound},
		{"товар исчез", 1, true, true, cart.ErrProductNotFound},
		{"товар есть", 1, true, false, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fake := newFake()
			service := cart.NewService(fake)
			if tc.populate {
				require.NoError(t, service.Add(context.Background(), 1, 2, 2))
			}
			if tc.remove {
				fake.mu.Lock()
				delete(fake.products, 2)
				fake.mu.Unlock()
			}
			fake.calls.Store(0)
			view, err := service.Get(context.Background(), tc.userID)
			if tc.want == nil {
				require.NoError(t, err)
				require.Equal(t, uint32(40), view.TotalPrice)
			} else {
				require.ErrorIs(t, err, tc.want)
				require.Empty(t, view.Items)
			}
			if !tc.populate {
				require.Zero(t, fake.calls.Load())
			}
		})
	}
}

func TestGetCartRequestsProductsInParallel(t *testing.T) {
	t.Parallel()
	fake := newFake()
	service := cart.NewService(fake)
	for _, sku := range []int64{9, 2, 5} {
		require.NoError(t, service.Add(context.Background(), 1, sku, 1))
	}
	fake.calls.Store(0)
	fake.maximum.Store(0)
	fake.mu.Lock()
	fake.deadlines = nil
	fake.mu.Unlock()
	fake.delay = 30 * time.Millisecond
	view, err := service.Get(context.Background(), 1)
	require.NoError(t, err)
	require.Greater(t, fake.maximum.Load(), int32(1))
	require.Equal(t, int32(3), fake.calls.Load())
	require.Equal(t, []int64{2, 5, 9}, []int64{view.Items[0].SKU, view.Items[1].SKU, view.Items[2].SKU})
	require.Equal(t, uint32(160), view.TotalPrice)
	fake.mu.RLock()
	require.Len(t, fake.deadlines, 3)
	require.Equal(t, fake.deadlines[0], fake.deadlines[1])
	require.Equal(t, fake.deadlines[1], fake.deadlines[2])
	fake.mu.RUnlock()
}

func TestCartHTTPContract(t *testing.T) {
	t.Parallel()
	fake := newFake()
	handler := cart.Handler(cart.NewService(fake), log.New(io.Discard, "", 0))
	for _, sku := range []string{"9", "2", "5"} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/user/42/cart/"+sku, strings.NewReader(`{"count":2}`)))
		require.Equal(t, http.StatusNoContent, recorder.Code)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/user/42/cart", nil))
	require.Equal(t, http.StatusOK, response.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	items, ok := body["items"].([]any)
	require.True(t, ok)
	require.Len(t, items, 3)
	require.Equal(t, float64(2), items[0].(map[string]any)["sku"])
	require.Equal(t, float64(5), items[1].(map[string]any)["sku"])
	require.Equal(t, float64(9), items[2].(map[string]any)["sku"])
	require.Equal(t, float64(2), items[0].(map[string]any)["count"])
	require.Equal(t, float64(20), items[0].(map[string]any)["price"])
	require.Equal(t, float64(320), body["total_price"])
	fake.mu.Lock()
	delete(fake.products, 5)
	fake.mu.Unlock()
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/user/42/cart", nil))
	require.Equal(t, http.StatusPreconditionFailed, response.Code)
	require.NotContains(t, response.Body.String(), "items")
}

func TestCartHTTPInvalidRequests(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		method string
		path   string
		body   string
		code   int
	}{
		{"нулевой пользователь", http.MethodGet, "/user/0/cart", "", 400},
		{"пустая корзина", http.MethodGet, "/user/5/cart", "", 404},
		{"нет товара", http.MethodPost, "/user/5/cart/7", `{"count":1}`, 412},
		{"нулевое количество", http.MethodPost, "/user/5/cart/2", `{"count":0}`, 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			handler := cart.Handler(cart.NewService(newFake()), log.New(io.Discard, "", 0))
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body)))
			require.Equal(t, tc.code, response.Code)
		})
	}
}

func BenchmarkAddToCart(b *testing.B) {
	fake := newFake()
	fake.products[2].Price = 0
	service := cart.NewService(fake)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := service.Add(context.Background(), 1, 2, 1); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetCart(b *testing.B) {
	service := cart.NewService(newFake())
	for _, sku := range []int64{2, 5, 9} {
		if err := service.Add(context.Background(), 1, sku, 1); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := service.Get(context.Background(), 1); err != nil {
			b.Fatal(err)
		}
	}
}
