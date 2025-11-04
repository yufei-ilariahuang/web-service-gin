package main

import (
	"context"
	"fmt"
	"os"
	"src/database"
	"src/handlers"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Initialize database backend (select by DB_BACKEND env var)
	backend := os.Getenv("DB_BACKEND") // "mysql" (default) or "dynamodb"
	if backend == "dynamodb" {
		// DynamoDB init
		region := os.Getenv("AWS_REGION")
		if region == "" {
			region = "us-east-1"
		}
		endpoint := os.Getenv("DYNAMO_ENDPOINT") // optional local endpoint
		if err := database.InitDynamoDB(context.Background(), region, endpoint); err != nil {
			panic(err)
		}
		// table name and shards via env
		table := os.Getenv("DYNAMO_TABLE")
		if table == "" {
			table = "shopping_carts"
		}
		// set store implementation
		database.SetStore(database.NewDynamoStore(table, 128))
	} else {
		// default to MySQL
		// Build DSN from environment variables
		var dsn string
		dbEndpoint := os.Getenv("DB_ENDPOINT")
		dbUser := os.Getenv("DB_USER")
		dbPassword := os.Getenv("DB_PASSWORD")
		dbName := os.Getenv("DB_NAME")

		if dbEndpoint == "" || dbUser == "" || dbPassword == "" || dbName == "" {
			// Local development fallback
			dsn = "admin:301752Qq!@tcp(terraform-20251103214459198300000001.cpkkycme8ops.us-west-2.rds.amazonaws.com:3306)/ecommerce"
		} else {
			dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s", dbUser, dbPassword, dbEndpoint, dbName)
		}

		if err := database.InitDB(dsn); err != nil {
			panic(err)
		}
		defer database.CloseDB()

		// Run migrations to ensure tables exist
		if err := database.RunMigrations(); err != nil {
			panic(fmt.Sprintf("failed to run migrations: %v", err))
		}

		database.SetStore(database.NewMySQLStore())
	}

	// Create Echo instance
	e := echo.New()

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status": "healthy",
		})
	})

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.POST("/shopping-carts", handlers.CreateShoppingCart)
	e.GET("/shopping-carts/:shoppingCartId", handlers.GetShoppingCart)
	e.POST("/shopping-carts/:shoppingCartId/items", handlers.AddItemsToCart)

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}
