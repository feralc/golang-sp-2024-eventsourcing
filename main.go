package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/feralc/golang-sp-2024-eventsourcing/api"
	"github.com/feralc/golang-sp-2024-eventsourcing/application/service"
	"github.com/feralc/golang-sp-2024-eventsourcing/domain/event"
	"github.com/feralc/golang-sp-2024-eventsourcing/esourcing"
	"github.com/feralc/golang-sp-2024-eventsourcing/infrastructure/persistence"
	"github.com/feralc/golang-sp-2024-eventsourcing/infrastructure/projection"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Invalid command")
	}

	godotenv.Load()

	eventstoreHost := os.Getenv("EVENTSTORE_HOST")
	eventstoreUsername := os.Getenv("EVENTSTORE_USER")
	eventstorePassword := os.Getenv("EVENTSTORE_PASS")

	store, err := esourcing.NewEventStore(esourcing.EventStoreConfig{
		Host:     eventstoreHost,
		Username: eventstoreUsername,
		Password: eventstorePassword,
	})

	if err != nil {
		log.Fatal(err)
	}

	store.RegisterEventType((*event.ShoppingCartCreated)(nil))
	store.RegisterEventType((*event.ShoppingCartItemAdded)(nil))
	store.RegisterEventType((*event.ShoppingCartItemRemoved)(nil))
	store.RegisterEventType((*event.ShoppingCartCheckedOut)(nil))

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME")),
	)
	if err != nil {
		panic(err)
	}

	defer db.Close()

	err = setupDatabase(db)
	if err != nil {
		log.Fatalf("Error setting up database: %v", err)
	}

	svc, err := service.New(db)

	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	cartRepository := persistence.NewEventSourcedShoppingCartRepository(store)
	productRepository := persistence.NewInMemoryProductRepository()
	shoppingCartService := service.NewShoppingCartService(cartRepository, productRepository)

	cmd := os.Args[1]

	switch cmd {

	case "start:server":
		e := echo.New()

		e.POST("/shopping-cart", api.CreateShoppingCartHandler(shoppingCartService))
		e.POST("/shopping-cart/:cartID/item", api.AddItemHandler(shoppingCartService))
		e.DELETE("/shopping-cart/:cartID/item/:productID", api.RemoveItemHandler(shoppingCartService))
		e.POST("/shopping-cart/:cartID/checkout", api.CheckoutHandler(shoppingCartService))
		e.GET("/shopping-cart/:cartID", api.GetShoppingCartHandler(cartRepository))
		e.GET("/shopping-carts", api.GetAllShoppingCartsHandler(db))

		e.Logger.Fatal(e.Start(":8080"))

	case "start:projection":
		personProjection := projection.NewShoppingCartProjection(svc, store)
		personProjection.Run(ctx)
	}
}

func setupDatabase(db *sql.DB) error {
	queries := []string{
		`DROP TABLE IF EXISTS es_subscription_checkpoint;`,
		`DROP TABLE IF EXISTS shopping_cart_item;`,
		`DROP TABLE IF EXISTS shopping_cart;`,
		`CREATE TABLE es_subscription_checkpoint (
			subscription_id VARCHAR(255) PRIMARY KEY,
			checkpoint_position VARCHAR(255),
			checkpoint_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE shopping_cart (
			cart_id VARCHAR(255) PRIMARY KEY,
			total DECIMAL(10,2) DEFAULT 0.0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE shopping_cart_item (
			cart_id VARCHAR(255) NOT NULL,
			product_id VARCHAR(255) NOT NULL,
			quantity INT NOT NULL,
			price DECIMAL(10,2) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (cart_id, product_id),
			FOREIGN KEY (cart_id) REFERENCES shopping_cart(cart_id)
		);`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			return fmt.Errorf("error executing query: %w", err)
		}
	}

	return nil
}
