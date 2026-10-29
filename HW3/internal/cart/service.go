package cart

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
	"sync/atomic"
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

type storedItem struct {
	count atomic.Uint32
	price uint32
}

type cartEntry struct {
	sku   int64
	count uint32
}

type Service struct {
	client ProductClient
	mu     sync.RWMutex
	carts  map[int64]map[int64]*storedItem
}

func NewService(client ProductClient) *Service {
	return &Service{client: client, carts: make(map[int64]map[int64]*storedItem)}
}

func (s *Service) product(ctx context.Context, sku int64) (*productv1.GetProductResponse, error) {
	response, err := s.client.GetProduct(ctx, &productv1.GetProductRequest{Sku: sku})
	if status.Code(err) == codes.NotFound {
		return nil, fmt.Errorf("проверить товар: %w", ErrProductNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("проверить товар через gRPC: %w", err)
	}
	if response == nil || response.GetSku() != sku || response.GetPrice() < 0 || response.GetName() == "" {
		return nil, fmt.Errorf("проверить товар: %w", ErrInvalidProduct)
	}
	return response, nil
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
	response, err := s.product(requestCtx, sku)
	if err != nil {
		return err
	}
	price := uint32(response.GetPrice())
	s.mu.Lock()
	defer s.mu.Unlock()
	userCart := s.carts[userID]
	current := userCart[sku]
	var previous uint32
	if current != nil {
		previous = current.count.Load()
	}
	if count > math.MaxUint32-previous {
		return ErrOverflow
	}
	updated := previous + count
	var total uint64
	for existingSKU, existing := range userCart {
		if existingSKU != sku {
			total += uint64(existing.count.Load()) * uint64(existing.price)
		}
	}
	total += uint64(updated) * uint64(price)
	if total > math.MaxUint32 {
		return ErrOverflow
	}
	if userCart == nil {
		userCart = make(map[int64]*storedItem)
		s.carts[userID] = userCart
	}
	if current == nil {
		current = &storedItem{}
		userCart[sku] = current
	}
	current.price = price
	current.count.Add(count)
	return nil
}

func (s *Service) Get(ctx context.Context, userID int64) (View, error) {
	if userID < 1 {
		return View{}, ErrInvalidUser
	}
	s.mu.RLock()
	userCart := s.carts[userID]
	entries := make([]cartEntry, 0, len(userCart))
	for sku, item := range userCart {
		entries = append(entries, cartEntry{sku: sku, count: item.count.Load()})
	}
	s.mu.RUnlock()
	if len(entries) == 0 {
		return View{}, ErrCartNotFound
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].sku < entries[j].sku })
	requestCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	items := make([]Item, len(entries))
	errorsByIndex := make([]error, len(entries))
	var group sync.WaitGroup
	for index, entry := range entries {
		group.Add(1)
		go func(index int, entry cartEntry) {
			defer group.Done()
			response, err := s.product(requestCtx, entry.sku)
			if err != nil {
				errorsByIndex[index] = err
				return
			}
			items[index] = Item{SKU: entry.sku, Name: response.GetName(), Count: entry.count, Price: uint32(response.GetPrice())}
		}(index, entry)
	}
	group.Wait()
	for _, err := range errorsByIndex {
		if err != nil {
			return View{}, err
		}
	}
	var total uint64
	for _, item := range items {
		total += uint64(item.Count) * uint64(item.Price)
		if total > math.MaxUint32 {
			return View{}, ErrOverflow
		}
	}
	return View{Items: items, TotalPrice: uint32(total)}, nil
}
