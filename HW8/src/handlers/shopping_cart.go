package handlers

import (
	"net/http"
	"src/api"
	"src/database"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

// CreateShoppingCart handles POST /shopping-carts
func CreateShoppingCart(c echo.Context) error {
	req := struct {
		CustomerID int32 `json:"customer_id"`
	}{}
	// Validation 1: Parse JSON (prevents malformed input)
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, api.Error{
			Error:   "INVALID_INPUT",
			Message: "Invalid request body",
		})
	}
	// Validation 2: Customer ID must be positive
	if req.CustomerID <= 0 {
		return c.JSON(http.StatusBadRequest, api.Error{
			Error:   "INVALID_INPUT",
			Message: "Customer ID must be positive",
		})
	}
	// Database call uses parameterized query (safe from injection)
	cartID, err := database.CreateCart(c.Request().Context(), req.CustomerID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.Error{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to create cart",
		})
	}

	return c.JSON(http.StatusCreated, map[string]int32{
		"shopping_cart_id": cartID,
	})
}

// GetShoppingCart handles GET /shopping-carts/{id}
func GetShoppingCart(c echo.Context) error {
	// Validation 1: Extract parameter safely
	cartIDStr := c.Param("shoppingCartId")
	// Validation 2: Remove whitespace (prevent bypasses)
	cartIDStr = strings.TrimSpace(cartIDStr)
	// Validation 3: Parse to integer
	cartID, err := strconv.ParseInt(cartIDStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, api.Error{
			Error:   "INVALID_INPUT",
			Message: "Invalid cart ID",
		})
	}
	// Validation 4: Ensure positive
	if cartID <= 0 {
		return c.JSON(http.StatusBadRequest, api.Error{
			Error:   "INVALID_INPUT",
			Message: "Cart ID must be positive",
		})
	}
	// Parameterized query (safe)
	cart, err := database.GetCart(c.Request().Context(), int32(cartID))
	if err != nil {
		return c.JSON(http.StatusNotFound, api.Error{
			Error:   "NOT_FOUND",
			Message: "Shopping cart not found",
		})
	}

	return c.JSON(http.StatusOK, cart)
}

// AddItemsToCart handles POST /shopping-carts/{id}/items
func AddItemsToCart(c echo.Context) error {
	// Validate cart ID from path
	cartIDStr := c.Param("shoppingCartId")
	cartID, err := strconv.ParseInt(cartIDStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, api.Error{
			Error:   "INVALID_INPUT",
			Message: "Invalid cart ID",
		})
	}
	// Validate request body
	req := struct {
		ProductID int32 `json:"product_id"`
		Quantity  int32 `json:"quantity"`
	}{}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, api.Error{
			Error:   "INVALID_INPUT",
			Message: "Invalid request body",
		})
	}
	// Validate product ID and quantity
	if req.ProductID <= 0 || req.Quantity <= 0 {
		return c.JSON(http.StatusBadRequest, api.Error{
			Error:   "INVALID_INPUT",
			Message: "Product ID and quantity must be positive",
		})
	}
	// Safe database operation with parameterized query
	err = database.AddItemToCart(c.Request().Context(), int32(cartID), req.ProductID, req.Quantity)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, api.Error{
				Error:   "NOT_FOUND",
				Message: "Cart or product not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, api.Error{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to add item",
		})
	}

	return c.NoContent(http.StatusNoContent)
}
