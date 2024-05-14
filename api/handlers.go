package api

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/feralc/golang-sp-2024-eventsourcing/application/service"
	"github.com/feralc/golang-sp-2024-eventsourcing/domain/repository"
	"github.com/labstack/echo/v4"
)

func CreateShoppingCartHandler(svc *service.ShoppingCartService) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := context.Background()
		cartID, err := svc.CreateShoppingCart(ctx)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusCreated, map[string]string{"cartID": cartID})
	}
}

func AddItemHandler(svc *service.ShoppingCartService) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := context.Background()
		cartID := c.Param("cartID")

		data := echo.Map{}
		if err := c.Bind(&data); err != nil {
			return err
		}

		productID := fmt.Sprintf("%v", data["product_id"])
		quantity, _ := strconv.Atoi(fmt.Sprintf("%v", data["quantity"]))

		err := svc.AddItem(ctx, cartID, productID, quantity)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.NoContent(http.StatusOK)
	}
}

func RemoveItemHandler(svc *service.ShoppingCartService) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := context.Background()
		cartID := c.Param("cartID")
		productID := c.Param("productID")

		err := svc.RemoveItem(ctx, cartID, productID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.NoContent(http.StatusOK)
	}
}

func CheckoutHandler(svc *service.ShoppingCartService) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := context.Background()
		cartID := c.Param("cartID")

		err := svc.Checkout(ctx, cartID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.NoContent(http.StatusOK)
	}
}

func GetShoppingCartHandler(cartRepo repository.ShoppingCartRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		cartID := c.Param("cartID")
		cart, err := cartRepo.FindByID(c.Request().Context(), cartID)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Shopping cart not found"})
		}

		return c.JSON(http.StatusOK, NewShoppingCartViewModel(cart))
	}
}

func GetAllShoppingCartsHandler(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		query := `
			SELECT
				c.cart_id,
				c.total,
				c.created_at,
				i.product_id,
				i.name,
				i.quantity,
				i.price
			FROM
				shopping_cart c
			LEFT JOIN
				shopping_cart_item i ON c.cart_id = i.cart_id
		`

		rows, err := db.Query(query)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error": "Failed to query database",
			})
		}
		defer rows.Close()

		carts := map[string]ShoppingCartViewModel{}
		for rows.Next() {
			var cartID string
			var total sql.NullFloat64
			var price sql.NullFloat64
			var quantity sql.NullInt32
			var createdAtCart string
			var productID sql.NullString
			var productName sql.NullString

			if err := rows.Scan(&cartID, &total, &createdAtCart, &productID, &productName, &quantity, &price); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]interface{}{
					"error": fmt.Sprintf("Failed to scan rows %s", err),
				})
			}

			_, ok := carts[cartID]
			if !ok {
				carts[cartID] = ShoppingCartViewModel{
					CartID: cartID,
					Total:  total.Float64,
					Items:  []ShoppingCartItemViewModel{},
				}
			}

			cart := carts[cartID]

			if productID.Valid {
				cart.Items = append(cart.Items, ShoppingCartItemViewModel{
					ProductID: productID.String,
					Name:      productName.String,
					Price:     price.Float64,
					Quantity:  int(quantity.Int32),
					Total:     price.Float64 * float64(quantity.Int32),
				})
				carts[cartID] = cart
			}
		}

		response := []ShoppingCartViewModel{}
		for _, cart := range carts {
			response = append(response, cart)
		}

		return c.JSON(http.StatusOK, response)
	}
}

func GetAllProductsHandler(productRepo repository.ProductRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		products, err := productRepo.All(c.Request().Context())
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error": "Failed to fetch products",
			})
		}

		return c.JSON(http.StatusOK, products)
	}
}
