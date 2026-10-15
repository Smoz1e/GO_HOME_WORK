package product_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	productv1 "gitlab.ozon.dev/go/classrom/HW2/gen/product/v1"
	"gitlab.ozon.dev/go/classrom/HW2/internal/product"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLoadFileValidatesProducts(t *testing.T) {
	path := filepath.Join(t.TempDir(), "products.json")
	if err := os.WriteFile(path, []byte(`{"5":{"name":"Пять","price":15},"2":{"name":"Два","price":7}}`), 0600); err != nil {
		t.Fatal(err)
	}
	products, err := product.LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(products) != 2 || products[5].Name != "Пять" || products[2].Price != 7 {
		t.Fatalf("unexpected products: %+v", products)
	}
	for _, input := range []string{
		`{"0":{"name":"Ноль","price":1}}`,
		`{"1":{"name":"","price":1}}`,
		`{"1":{"name":"Слишком дорого","price":2147483648}}`,
		`{"1":{"name":"Отрицательная цена","price":-1}}`,
	} {
		if err := os.WriteFile(path, []byte(input), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := product.LoadFile(path); err == nil {
			t.Fatalf("accepted invalid products: %s", input)
		}
	}
}

func TestGetProductCodesAndListPagination(t *testing.T) {
	s := product.NewServer(map[int64]product.Product{
		9: {SKU: 9, Name: "Девять", Price: 90},
		2: {SKU: 2, Name: "Два", Price: 20},
		5: {SKU: 5, Name: "Пять", Price: 50},
	})
	ctx := context.Background()
	for _, sku := range []int64{0, -1} {
		_, err := s.GetProduct(ctx, &productv1.GetProductRequest{Sku: sku})
		if status.Code(err) != codes.InvalidArgument {
			t.Fatalf("sku %d: got %v", sku, err)
		}
	}
	if _, err := s.GetProduct(ctx, &productv1.GetProductRequest{Sku: 3}); status.Code(err) != codes.NotFound {
		t.Fatalf("missing SKU: %v", err)
	}
	got, err := s.GetProduct(ctx, &productv1.GetProductRequest{Sku: 5})
	if err != nil {
		t.Fatal(err)
	}
	if got.GetSku() != 5 || got.GetName() != "Пять" || got.GetPrice() != 50 {
		t.Fatalf("unexpected product: %+v", got)
	}
	list, err := s.ListProducts(ctx, &productv1.ListProductsRequest{Page: 2, Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if list.GetTotal() != 3 || list.GetTotalPages() != 2 || len(list.GetProducts()) != 1 || list.GetProducts()[0].GetSku() != 9 {
		t.Fatalf("unexpected second page: %+v", list)
	}
	if _, err := s.ListProducts(ctx, &productv1.ListProductsRequest{Page: 0, Limit: 2}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("invalid pagination: %v", err)
	}
}
