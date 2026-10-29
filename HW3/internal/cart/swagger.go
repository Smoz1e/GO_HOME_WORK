package cart

const openAPISpec = `openapi: 3.0.3
info:
  title: API корзины
  version: 1.0.0
servers:
  - url: http://localhost:8080
paths:
  /user/{user_id}/cart/{sku_id}:
    post:
      summary: Добавить товар в корзину
      parameters:
        - name: user_id
          in: path
          required: true
          schema: {type: integer, format: int64, minimum: 1}
        - name: sku_id
          in: path
          required: true
          schema: {type: integer, format: int64, minimum: 1}
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [count]
              properties:
                count: {type: integer, format: int64, minimum: 1, maximum: 4294967295}
      responses:
        '204': {description: Товар добавлен}
        '400': {description: Некорректный пользователь, SKU или количество}
        '412': {description: Товар не существует}
        '500': {description: Внутренняя ошибка}
  /user/{user_id}/cart:
    get:
      summary: Получить корзину пользователя
      parameters:
        - name: user_id
          in: path
          required: true
          schema: {type: integer, format: int64, minimum: 1}
      responses:
        '200':
          description: Содержимое корзины
          content:
            application/json:
              schema:
                type: object
                required: [items, total_price]
                properties:
                  items:
                    type: array
                    items:
                      type: object
                      required: [sku, name, count, price]
                      properties:
                        sku: {type: integer, format: int64}
                        name: {type: string}
                        count: {type: integer, format: int64, minimum: 1, maximum: 4294967295}
                        price: {type: integer, format: int32, minimum: 0}
                  total_price: {type: integer, format: int64, minimum: 0, maximum: 4294967295}
        '400': {description: Некорректный идентификатор пользователя}
        '404': {description: Корзина отсутствует или пуста}
`

const swaggerHTML = `<!doctype html>
<html lang="ru"><head><meta charset="utf-8"><title>Cart API — Swagger</title>
<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"></head>
<body><div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<script>SwaggerUIBundle({url:'/swagger/openapi.yaml',dom_id:'#swagger-ui'});</script>
</body></html>`
