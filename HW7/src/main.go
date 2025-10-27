package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/labstack/echo/v4"
)

// Product defines model for Product.
type Product struct {
	Brand       string  `json:"brand"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Price       float32 `json:"price"`
}

// ProductInput defines model for ProductInput.
type ProductInput struct {
	Brand       *string `json:"brand,omitempty"`
	Category    string  `json:"category"`
	Description *string `json:"description,omitempty"`
	Name        string  `json:"name"`
	Price       float32 `json:"price"`
}

// GetProductsParams defines parameters for GetProducts.
type GetProductsParams struct {
	Brands     *string `form:"brands,omitempty" json:"brands,omitempty"`
	Categories *string `form:"categories,omitempty" json:"categories,omitempty"`
}

// Response defines the structure for API responses.
type Response struct {
	Products   []Product `json:"products"`
	TotalFound int       `json:"total_found"`
	SearchTime string    `json:"search_time"`
}

var (
	products []Product

	// Define specific brands and categories
	brands     = []string{"Apple", "Samsung", "Google", "Microsoft", "Sony", "Logitech"}
	categories = []string{"Electronics", "Books", "Home", "Clothing", "Groceries", "Toys"}
)

func main() {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Generate initial product data
	generateProducts(100000)

	e := echo.New()

	// Routes
	e.GET("/products", getProducts)
	e.POST("/products", createProduct)
	e.DELETE("/products/:id", deleteProductById)
	e.GET("/products/:id", getProductById)

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}

// generateProducts creates a specified number of products with fake data.
func generateProducts(count int) {
	products = make([]Product, count)
	for i := 0; i < count; i++ {
		// Randomly select brand and category from defined slices
		brand := brands[rand.Intn(len(brands))]
		category := categories[rand.Intn(len(categories))]

		productName := fmt.Sprintf("Product %s %d", gofakeit.AppName(), i+1)
		description := gofakeit.BS()
		price := rand.Float32() * 1000 // Random price between 0 and 1000

		products[i] = Product{
			ID:          strconv.Itoa(i + 1),
			Name:        productName,
			Category:    category,
			Brand:       brand,
			Description: description,
			Price:       price,
		}
	}
	fmt.Printf("Generated %d products.\n", count)
}

// getProducts handles requests to GET /products.
func getProducts(c echo.Context) error {
	startTime := time.Now()

	// --- Search Parameters ---
	brandsParam := c.QueryParam("brands")
	categoriesParam := c.QueryParam("categories")

	var filteredProducts []Product
	checkedCount := 0 // This counter is for tracking how many products were checked.

	// Iterate through a limited number of products for searching
	limit := 100
	if len(products) < limit {
		limit = len(products)
	}

	for i := 0; i < limit; i++ {
		checkedCount++ // Increment for every product checked
		product := products[i]
		match := true

		// Filter by brands (case-insensitive, partial match)
		if brandsParam != "" {
			if !strings.Contains(strings.ToLower(product.Brand), strings.ToLower(brandsParam)) {
				match = false
			}
		}

		// Filter by categories (case-insensitive, exact match)
		if match && categoriesParam != "" {
			if !strings.EqualFold(product.Category, categoriesParam) {
				match = false
			}
		}

		if match {
			filteredProducts = append(filteredProducts, product)
		}
	}

	// --- Response ---
	elapsedTime := time.Since(startTime).Round(time.Millisecond)

	// Limit results to 20
	if len(filteredProducts) > 20 {
		filteredProducts = filteredProducts[:20]
	}

	response := Response{
		Products:   filteredProducts,
		TotalFound: len(filteredProducts), // In this simulation, this is the count of matches within the checked limit.
		SearchTime: elapsedTime.String(),
	}

	return c.JSON(http.StatusOK, response)
}

// createProduct handles requests to POST /products.
func createProduct(c echo.Context) error {
	productInput := new(ProductInput)
	if err := c.Bind(productInput); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Generate a new unique ID
	// For simplicity in this basic version, we'll assume IDs are sequential.
	// In a real-world scenario, you'd need a more robust ID generation mechanism.
	newID := len(products) + 1

	newProduct := Product{
		ID:          strconv.Itoa(newID),
		Name:        productInput.Name,
		Category:    productInput.Category,
		Price:       productInput.Price,
		Description: "No description provided", // Default description
		Brand:       "Unknown Brand",           // Default brand
	}

	if productInput.Description != nil {
		newProduct.Description = *productInput.Description
	}
	if productInput.Brand != nil {
		newProduct.Brand = *productInput.Brand
	}

	// Append the new product directly without locks
	products = append(products, newProduct)

	return c.JSON(http.StatusCreated, newProduct)
}

// deleteProductById handles requests to DELETE /products/{id}.
func deleteProductById(c echo.Context) error {
	id := c.Param("id")

	// Find the product and remove it without locks
	for i, p := range products {
		if p.ID == id {
			// Remove the product by slicing
			products = append(products[:i], products[i+1:]...)
			return c.NoContent(http.StatusNoContent)
		}
	}

	return echo.NewHTTPError(http.StatusNotFound, "Product not found")
}

// getProductById handles requests to GET /products/{id}.
func getProductById(c echo.Context) error {
	id := c.Param("id")

	// Retrieve the product without locks
	for _, p := range products {
		if p.ID == id {
			return c.JSON(http.StatusOK, p)
		}
	}

	return echo.NewHTTPError(http.StatusNotFound, "Product not found")
}
