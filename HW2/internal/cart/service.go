package cart

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	productv1 "gitlab.ozon.dev/go/classrom/HW2/gen/product/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInvalidUser     = errors.New("ID пользователя должен быть натуральным числом")
	ErrInvalidSKU      = errors.New("SKU должен быть натуральным числом")
	ErrInvalidCount    = errors.New("количество должно быть натуральным числом")
	ErrProductNotFound = errors.New("товар не найден в сервисе товаров")
	ErrCartNotFound    = errors.New("корзина не найдена")
	ErrOverflow        = errors.New("сумма или количество товаров превышает допустимое значение")
	ErrInvalidProduct  = errors.New("сервис товаров вернул некорректные данные")
)

type ProductClient interface {
	GetProduct(context.Context, *productv1.GetProductRequest, ...grpc.CallOption) (*productv1.GetProductResponse, error)
}

type Item struct {
	SKU   int64  `json:"sku"`
	Name  string `json:"name"`
	Count uint32 `json:"count"`
	Price uint32 `json:"price"`
}

type View struct {
	Items      []Item `json:"items"`
	TotalPrice uint32 `json:"total_price"`
}

type Service struct {
	client ProductClient
	mu     sync.RWMutex
	carts  map[int64]map[int64]Item
}

func NewService(client ProductClient) *Service {
	return &Service{client: client, carts: make(map[int64]map[int64]Item)}
}

func (s *Service) Add(ctx context.Context, userID, sku int64, count uint32) error {
	if userID < 1 {
		return ErrInvalidUser
	}
	if sku < 1 {
		return ErrInvalidSKU
	}
	if count == 0 {
		return ErrInvalidCount
	}
	requestCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	response, err := s.client.GetProduct(requestCtx, &productv1.GetProductRequest{Sku: sku})
	if status.Code(err) == codes.NotFound {
		return fmt.Errorf("проверить товар: %w", ErrProductNotFound)
	}
	if err != nil {
		return fmt.Errorf("проверить товар через gRPC: %w", err)
	}
	if response == nil || response.GetSku() != sku || response.GetPrice() < 0 || response.GetName() == "" {
		return fmt.Errorf("проверить товар: %w", ErrInvalidProduct)
	}
	item := Item{SKU: sku, Name: response.GetName(), Count: count, Price: uint32(response.GetPrice())}
	s.mu.Lock()
	defer s.mu.Unlock()
	userCart := s.carts[userID]
	if current, exists := userCart[sku]; exists {
		if count > math.MaxUint32-current.Count {
			return ErrOverflow
		}
		item.Count += current.Count
	}
	var total uint64
	for existingSKU, existing := range userCart {
		if existingSKU == sku {
			continue
		}
		total += uint64(existing.Count) * uint64(existing.Price)
		if total > math.MaxUint32 {
			return ErrOverflow
		}
	}
	total += uint64(item.Count) * uint64(item.Price)
	if total > math.MaxUint32 {
		return ErrOverflow
	}
	if userCart == nil {
		userCart = make(map[int64]Item)
		s.carts[userID] = userCart
	}
	userCart[sku] = item
	return nil
}

func (s *Service) Get(userID int64) (View, error) {
	if userID < 1 {
		return View{}, ErrInvalidUser
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	userCart := s.carts[userID]
	if len(userCart) == 0 {
		return View{}, ErrCartNotFound
	}
	view := View{Items: make([]Item, 0, len(userCart))}
	var total uint64
	for _, item := range userCart {
		view.Items = append(view.Items, item)
		total += uint64(item.Count) * uint64(item.Price)
	}
	sort.Slice(view.Items, func(i, j int) bool { return view.Items[i].SKU < view.Items[j].SKU })
	view.TotalPrice = uint32(total)
	return view, nil
}
