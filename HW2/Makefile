.PHONY: run-all generate test

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
	@mkdir -p gen
	@if command -v protoc >/dev/null 2>&1; then \
		protoc -I api --go_out=gen --go_opt=paths=source_relative \
			--go-grpc_out=gen --go-grpc_opt=paths=source_relative \
			api/product/v1/product.proto; \
	elif command -v buf >/dev/null 2>&1; then \
		buf generate; \
	else \
		echo 'Установите protoc или buf и плагины protoc-gen-go, protoc-gen-go-grpc' >&2; \
		exit 1; \
		fi
	@go run ./tools/stripcomments gen/product/v1/product.pb.go gen/product/v1/product_grpc.pb.go

test:
	go test ./...
