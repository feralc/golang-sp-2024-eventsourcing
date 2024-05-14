# Event sourcing with DDD Aggregates and CQRS

```bash
docker-compose up
```

```bash
go run main.go start:server
```

```bash
go run main.go start:projection
```

## API Curl Commands

### Create Shopping Cart

```bash
curl -X POST -H "Content-Type: application/json" http://localhost:8080/shopping-cart
```

### Add Item to Shopping Cart

```bash
curl -X POST -H "Content-Type: application/json" -d '{"product_id":"123", "quantity":2}' http://localhost:8080/shopping-cart/364ae8b5-95e6-4c32-bbb0-1d0449d17814/item
```

### Remove Item from Shopping Cart

```bash
curl -X DELETE http://localhost:8080/shopping-cart/364ae8b5-95e6-4c32-bbb0-1d0449d17814/item/123
```

### Checkout Shopping Cart

```bash
curl -X POST -H "Content-Type: application/json" http://localhost:8080/shopping-cart/364ae8b5-95e6-4c32-bbb0-1d0449d17814/checkout
```
