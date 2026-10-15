package cart_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	productv1 "gitlab.ozon.dev/go/classrom/HW2/gen/product/v1"
	"gitlab.ozon.dev/go/classrom/HW2/internal/cart"
	"gitlab.ozon.dev/go/classrom/HW2/internal/product"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func testHandler(t *testing.T, logOut *bytes.Buffer) http.Handler {
	t.Helper()
	lis := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer()
	productv1.RegisterProductServiceServer(grpcServer, product.NewServer(map[int64]product.Product{
		8: {SKU: 8, Name: "Восемь", Price: 12},
		3: {SKU: 3, Name: "Три", Price: 7},
	}))
	go func() { _ = grpcServer.Serve(lis) }()
	t.Cleanup(grpcServer.Stop)
	conn, err := grpc.NewClient("passthrough:///bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	client := productv1.NewProductServiceClient(conn)
	probeCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := client.GetProduct(probeCtx, &productv1.GetProductRequest{Sku: 8}); err != nil {
		t.Fatalf("test gRPC connection: %v", err)
	}
	return cart.Handler(cart.NewService(client), log.New(logOut, "", 0))
}

func request(handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	r := httptest.NewRecorder()
	handler.ServeHTTP(r, httptest.NewRequest(method, path, strings.NewReader(body)))
	return r
}

func TestCartUsesGRPCAndReturnsSortedAggregatedItems(t *testing.T) {
	var logs bytes.Buffer
	h := testHandler(t, &logs)
	for _, item := range []struct{ path, body string }{
		{"/user/42/cart/8", `{"count":2}`},
		{"/user/42/cart/3", `{"count":1}`},
		{"/user/42/cart/8", `{"count":3}`},
	} {
		resp := request(h, http.MethodPost, item.path, item.body)
		if resp.Code != http.StatusNoContent || resp.Body.Len() != 0 {
			t.Fatalf("POST %s: status=%d body=%s", item.path, resp.Code, resp.Body.String())
		}
	}
	resp := request(h, http.MethodGet, "/user/42/cart", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("GET: %d %s", resp.Code, resp.Body.String())
	}
	var got struct {
		Items []struct {
			SKU   int64  `json:"sku"`
			Name  string `json:"name"`
			Count uint32 `json:"count"`
			Price uint32 `json:"price"`
		} `json:"items"`
		TotalPrice uint32 `json:"total_price"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 2 || got.Items[0].SKU != 3 || got.Items[1].SKU != 8 || got.Items[1].Count != 5 || got.Items[1].Name != "Восемь" || got.Items[1].Price != 12 || got.TotalPrice != 67 {
		t.Fatalf("unexpected cart: %+v", got)
	}
	if !strings.Contains(logs.String(), "POST /user/42/cart/8 204") {
		t.Fatalf("request not logged: %q", logs.String())
	}
}

func TestCartHTTPValidationAndSwagger(t *testing.T) {
	var logs bytes.Buffer
	h := testHandler(t, &logs)
	for _, item := range []struct {
		method, path, body string
		code               int
	}{
		{"POST", "/user/0/cart/3", `{"count":1}`, 400},
		{"POST", "/user/1/cart/0", `{"count":1}`, 400},
		{"POST", "/user/1/cart/3", `{"count":0}`, 400},
		{"POST", "/user/1/cart/3", `{"count":-1}`, 400},
		{"POST", "/user/1/cart/9", `{"count":1}`, 412},
		{"GET", "/user/1/cart", "", 404},
		{"GET", "/user/0/cart", "", 400},
		{"GET", "/swagger/", "", 200},
		{"GET", "/swagger/openapi.yaml", "", 200},
	} {
		resp := request(h, item.method, item.path, item.body)
		if resp.Code != item.code {
			t.Fatalf("%s %s: got %d, want %d, body=%s", item.method, item.path, resp.Code, item.code, resp.Body.String())
		}
		if item.code >= 400 && !strings.Contains(resp.Body.String(), "ошибка") {
			t.Fatalf("missing Russian error: %s", resp.Body.String())
		}
	}
}
