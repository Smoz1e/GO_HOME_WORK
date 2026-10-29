.PHONY: run-all generate test test-race test-api bench

run-all:
	@set -eu; \
	workdir=$$(mktemp -d); \
	trap 'kill $${product_pid:-} $${cart_pid:-} 2>/dev/null || true; wait 2>/dev/null || true; rm -rf "$$workdir"' EXIT INT TERM; \
	go build -o "$$workdir/product-service" ./cmd/product-service; \
	go build -o "$$workdir/cart" ./cmd/cart; \
	"$$workdir/product-service" & product_pid=$$!; \
	"$$workdir/cart" & cart_pid=$$!; \
	while kill -0 $$product_pid 2>/dev/null && kill -0 $$cart_pid 2>/dev/null; do sleep 0.2; done; \
	if ! kill -0 $$product_pid 2>/dev/null; then \
		wait $$product_pid || true; \
		echo 'Сервис товаров завершился; останавливаем корзину' >&2; \
		exit 1; \
	fi; \
	wait $$cart_pid

generate:
	$(MAKE) -C ../HW2 generate

test:
	go test ./...

test-race:
	go test -race -count=1 ./... ./testdata/race

test-api:
	go test -v -count=1 ./testdata/api

bench:
	go test -run '^$$' -bench 'Benchmark(GetProduct|AddToCart|GetCart)$$' -benchmem ./internal/product ./internal/cart
