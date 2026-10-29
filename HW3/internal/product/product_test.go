package product_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	productv1 "gitlab.ozon.dev/go/classrom/HW2/gen/product/v1"
	"gitlab.ozon.dev/go/classrom/HW3/internal/product"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func products() map[int64]product.Product {
	return map[int64]product.Product{
		2: {SKU: 2, Name: "Два", Price: 20},
		5: {SKU: 5, Name: "Пять", Price: 50},
		9: {SKU: 9, Name: "Девять", Price: 90},
	}
}

func TestGetProduct(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		sku  int64
		code codes.Code
	}{
		{"существующий товар", 5, codes.OK},
		{"нулевой SKU", 0, codes.InvalidArgument},
		{"отрицательный SKU", -1, codes.InvalidArgument},
		{"неизвестный SKU", 7, codes.NotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			server := product.NewServer(products())
			got, err := server.GetProduct(context.Background(), &productv1.GetProductRequest{Sku: tc.sku})
			require.Equal(t, tc.code, status.Code(err))
			if tc.code == codes.OK {
				require.Equal(t, int64(5), got.GetSku())
				require.Equal(t, "Пять", got.GetName())
				require.Equal(t, int32(50), got.GetPrice())
			}
		})
	}
}

func TestListProducts(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		page  int32
		limit int32
		code  codes.Code
		skus  []int64
	}{
		{"первая страница", 1, 2, codes.OK, []int64{2, 5}},
		{"последняя страница", 2, 2, codes.OK, []int64{9}},
		{"страница за пределами", 3, 2, codes.OK, nil},
		{"нулевая страница", 0, 2, codes.InvalidArgument, nil},
		{"нулевой размер", 1, 0, codes.InvalidArgument, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			server := product.NewServer(products())
			got, err := server.ListProducts(context.Background(), &productv1.ListProductsRequest{Page: tc.page, Limit: tc.limit})
			require.Equal(t, tc.code, status.Code(err))
			if tc.code == codes.OK {
				require.Equal(t, int32(3), got.GetTotal())
				require.Equal(t, int32(2), got.GetTotalPages())
				var skus []int64
				for _, item := range got.GetProducts() {
					skus = append(skus, item.GetSku())
				}
				require.Equal(t, tc.skus, skus)
			}
		})
	}
}

func TestLoadFile(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		content string
		valid   bool
	}{
		{"корректный файл", `{"2":{"name":"Два","price":20}}`, true},
		{"нулевой SKU", `{"0":{"name":"Ноль","price":1}}`, false},
		{"пустое название", `{"2":{"name":"","price":1}}`, false},
		{"отрицательная цена", `{"2":{"name":"Два","price":-1}}`, false},
		{"цена вне int32", `{"2":{"name":"Два","price":2147483648}}`, false},
		{"не объект", `[]`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "products.json")
			require.NoError(t, os.WriteFile(path, []byte(tc.content), 0600))
			got, err := product.LoadFile(path)
			if tc.valid {
				require.NoError(t, err)
				require.Equal(t, uint32(20), got[2].Price)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func BenchmarkGetProduct(b *testing.B) {
	server := product.NewServer(products())
	request := &productv1.GetProductRequest{Sku: 5}
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := server.GetProduct(ctx, request); err != nil {
			b.Fatal(err)
		}
	}
}
