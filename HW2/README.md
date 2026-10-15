# Домашнее задание №2: товары и корзина

Из этой папки запустите оба сервиса:

```bash
make run-all
```

`product-service` слушает gRPC на `:50051` и при старте читает `products.json`. `cart` слушает HTTP на `:8080`. Корзины находятся в памяти и очищаются после перезапуска. Порты и файл можно изменить переменными `PRODUCT_GRPC_ADDR`, `CART_HTTP_ADDR` и `PRODUCTS_FILE`.

Пример работы:

```bash
curl -i -X POST http://localhost:8080/user/1007/cart/2958025 \
  -H 'Content-Type: application/json' -d '{"count":2}'
curl http://localhost:8080/user/1007/cart
```

Повторное добавление того же SKU увеличивает количество. Пустая корзина возвращает HTTP 404, несуществующий товар при добавлении — HTTP 412. Товары в ответе отсортированы по SKU. Описание HTTP API доступно по адресу `http://localhost:8080/swagger/`, спецификация — по `http://localhost:8080/swagger/openapi.yaml`. Для отображения Swagger UI странице нужен доступ к CDN; сам YAML доступен локально.

Контракт gRPC находится в `api/product/v1/product.proto`. Генерация Go-кода:

```bash
make generate
```

Нужны `protoc` и плагины `protoc-gen-go`, `protoc-gen-go-grpc` в `PATH`; если установлен `buf`, он используется как запасной вариант. Сгенерированные Go-файлы уже включены в проект.

Проверки из корня репозитория:

```bash
go test ./...
go build ./...
go vet ./...
```
