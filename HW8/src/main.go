package main

import (
	"os"
	"src/database"
	"src/handlers"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Initialize database
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "root:301752Qq!@tcp(localhost:3306)/ecommerce"
	}

	if err := database.InitDB(dsn); err != nil {
		panic(err)
	}
	defer database.CloseDB()

	// Create Echo instance
	e := echo.New()

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
