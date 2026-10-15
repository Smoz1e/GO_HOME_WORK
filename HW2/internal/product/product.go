package product

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	productv1 "gitlab.ozon.dev/go/classrom/HW2/gen/product/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Product struct {
	SKU   int64
	Name  string
	Price uint32
}

func LoadFile(path string) (map[int64]Product, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("прочитать файл товаров: %w", err)
	}
	var records map[string]json.RawMessage
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("разобрать файл товаров: %w", err)
	}
	if records == nil {
		return nil, errors.New("файл товаров должен содержать объект JSON")
	}
	products := make(map[int64]Product, len(records))
	for key, raw := range records {
		sku, err := strconv.ParseInt(key, 10, 64)
		if err != nil || sku < 1 {
			return nil, fmt.Errorf("некорректный SKU %q", key)
		}
		if _, exists := products[sku]; exists {
			return nil, fmt.Errorf("повторяющийся SKU %d", sku)
		}
		var record struct {
			Name  string `json:"name"`
			Price uint32 `json:"price"`
		}
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&record); err != nil {
			return nil, fmt.Errorf("разобрать товар %d: %w", sku, err)
		}
		if strings.TrimSpace(record.Name) == "" {
			return nil, fmt.Errorf("товар %d: название не может быть пустым", sku)
		}
		if record.Price > math.MaxInt32 {
			return nil, fmt.Errorf("товар %d: цена не помещается в int32", sku)
		}
		products[sku] = Product{SKU: sku, Name: record.Name, Price: record.Price}
	}
	return products, nil
}

type Server struct {
	productv1.UnimplementedProductServiceServer
	products map[int64]Product
}

func NewServer(products map[int64]Product) *Server {
	copyProducts := make(map[int64]Product, len(products))
	for sku, item := range products {
		copyProducts[sku] = item
	}
	return &Server{products: copyProducts}
}

func (s *Server) GetProduct(_ context.Context, request *productv1.GetProductRequest) (*productv1.GetProductResponse, error) {
	if request == nil || request.GetSku() < 1 {
		return nil, status.Error(codes.InvalidArgument, "SKU должен быть натуральным числом")
	}
	item, exists := s.products[request.GetSku()]
	if !exists {
		return nil, status.Error(codes.NotFound, "товар не найден")
	}
	if item.Price > math.MaxInt32 {
		return nil, status.Error(codes.Internal, "цена товара не помещается в int32")
	}
	return &productv1.GetProductResponse{Sku: item.SKU, Name: item.Name, Price: int32(item.Price)}, nil
}

func (s *Server) ListProducts(_ context.Context, request *productv1.ListProductsRequest) (*productv1.ListProductsResponse, error) {
	if request == nil || request.GetPage() < 1 || request.GetLimit() < 1 {
		return nil, status.Error(codes.InvalidArgument, "номер страницы и размер страницы должны быть положительными")
	}
	skus := make([]int64, 0, len(s.products))
	for sku := range s.products {
		skus = append(skus, sku)
	}
	sort.Slice(skus, func(i, j int) bool { return skus[i] < skus[j] })
	limit := int64(request.GetLimit())
	total := int64(len(skus))
	pages := total / limit
	if total%limit != 0 {
		pages++
	}
	response := &productv1.ListProductsResponse{Total: int32(total), TotalPages: int32(pages), Page: request.GetPage()}
	if int64(request.GetPage()) > pages {
		return response, nil
	}
	start := (int64(request.GetPage()) - 1) * limit
	end := start + limit
	if end > total {
		end = total
	}
	for _, sku := range skus[start:end] {
		item := s.products[sku]
		if item.Price > math.MaxInt32 {
			return nil, status.Error(codes.Internal, "цена товара не помещается в int32")
		}
		response.Products = append(response.Products, &productv1.Product{Sku: item.SKU, Name: item.Name, Price: int32(item.Price)})
	}
	return response, nil
}
