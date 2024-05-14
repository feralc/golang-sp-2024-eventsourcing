# API Curl Commands

### Create Shopping Cart

```bash
curl -X POST -H "Content-Type: application/json" http://localhost:8080/shopping-cart
```

### Add Item to Shopping Cart

```bash
curl -X POST -H "Content-Type: application/json" -d '{"product_id":"123", "quantity":2}' http://localhost:8080/shopping-cart/d541419a-9ee0-4f45-b1f6-24466abf5f88/item
```

### Remove Item from Shopping Cart

```bash
curl -X DELETE http://localhost:8080/shopping-cart/d541419a-9ee0-4f45-b1f6-24466abf5f88/item/123
```

### Checkout Shopping Cart

```bash
curl -X POST -H "Content-Type: application/json" http://localhost:8080/shopping-cart/d541419a-9ee0-4f45-b1f6-24466abf5f88/checkout
```
