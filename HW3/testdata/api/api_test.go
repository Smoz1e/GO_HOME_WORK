package api

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	productv1 "gitlab.ozon.dev/go/classrom/HW2/gen/product/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func randomUserID(t *testing.T) int64 {
	t.Helper()
	var data [8]byte
	_, err := rand.Read(data[:])
	require.NoError(t, err)
	return int64(binary.BigEndian.Uint64(data[:])%uint64(math.MaxInt64-1)) + 1
}

func callHTTP(t *testing.T, client *http.Client, method, path, body string) (int, []byte) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, method, cartURL()+path, strings.NewReader(body))
	require.NoError(t, err)
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := client.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()
	content, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	return response.StatusCode, content
}

func TestLiveServices(t *testing.T) {
	t.Parallel()
	products, conn, err := connectProduct()
	require.NoError(t, err)
	defer conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	item, err := products.GetProduct(ctx, &productv1.GetProductRequest{Sku: 2958025})
	require.NoError(t, err)
	require.Equal(t, int32(1028), item.GetPrice())
	_, err = products.GetProduct(ctx, &productv1.GetProductRequest{Sku: 0})
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	_, err = products.GetProduct(ctx, &productv1.GetProductRequest{Sku: 999999999})
	require.Equal(t, codes.NotFound, status.Code(err))
	userID := randomUserID(t)
	path := fmt.Sprintf("/user/%d/cart", userID)
	client := &http.Client{Timeout: 3 * time.Second}
	code, _ := callHTTP(t, client, http.MethodGet, path, "")
	require.Equal(t, http.StatusNotFound, code)
	code, _ = callHTTP(t, client, http.MethodPost, "/user/0/cart/2958025", `{"count":1}`)
	require.Equal(t, http.StatusBadRequest, code)
	code, _ = callHTTP(t, client, http.MethodPost, path+"/999999999", `{"count":1}`)
	require.Equal(t, http.StatusPreconditionFailed, code)
	code, _ = callHTTP(t, client, http.MethodPost, path+"/773297411", `{"count":1}`)
	require.Equal(t, http.StatusNoContent, code)
	code, _ = callHTTP(t, client, http.MethodPost, path+"/2958025", `{"count":2}`)
	require.Equal(t, http.StatusNoContent, code)
	code, content := callHTTP(t, client, http.MethodGet, path, "")
	require.Equal(t, http.StatusOK, code)
	var cart struct {
		Items []struct {
			SKU   int64  `json:"sku"`
			Name  string `json:"name"`
			Count uint32 `json:"count"`
			Price uint32 `json:"price"`
		} `json:"items"`
		TotalPrice uint32 `json:"total_price"`
	}
	require.NoError(t, json.Unmarshal(content, &cart))
	require.Len(t, cart.Items, 2)
	require.Equal(t, int64(2958025), cart.Items[0].SKU)
	require.Equal(t, uint32(2), cart.Items[0].Count)
	require.Equal(t, uint32(1028), cart.Items[0].Price)
	require.NotEmpty(t, cart.Items[0].Name)
	require.Equal(t, int64(773297411), cart.Items[1].SKU)
	require.Equal(t, uint32(4258), cart.TotalPrice)
}
