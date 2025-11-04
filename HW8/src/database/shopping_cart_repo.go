package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"src/api"
	"strings"
	"time"
)

// mysqlStore implements Store using the existing SQL schema/connection
type mysqlStore struct{}

func NewMySQLStore() Store {
	return &mysqlStore{}
}

// CreateCart creates a new shopping cart (MySQL)
func (m *mysqlStore) CreateCart(ctx context.Context, customerID int32) (int32, error) {
	query := "INSERT INTO shopping_carts (customer_id) VALUES (?)"

	result, err := DB.ExecContext(ctx, query, customerID)
	if err != nil {
		return 0, fmt.Errorf("failed to create cart: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int32(id), nil
}

// GetCart retrieves a shopping cart with all items (MySQL)
func (m *mysqlStore) GetCart(ctx context.Context, cartID int32) (*api.ShoppingCart, error) {
	query := `
		SELECT sc.id, sc.customer_id, sc.created_at
		FROM shopping_carts sc
		WHERE sc.id = ?
	`

	cart := &api.ShoppingCart{}
	var createdAtStr string // ← Parse as string first
	err := DB.QueryRowContext(ctx, query, cartID).Scan(&cart.ShoppingCartId, &cart.CustomerId, &createdAtStr)
	// Distinguish between different errors
	fmt.Printf("DEBUG: Query error: %v\n", err)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Specific: Cart doesn't exist
			return nil, fmt.Errorf("cart not found")
		}

		if errors.Is(err, context.DeadlineExceeded) {
			// Specific: Query timeout
			return nil, fmt.Errorf("database query timeout")
		}

		if strings.Contains(err.Error(), "connection") {
			// Specific: Connection issue
			return nil, fmt.Errorf("database connection failed")
		}

		// Generic database error
		return nil, fmt.Errorf("failed to retrieve cart: %w", err)
	}
	// Convert string to time.Time
	if createdAtStr != "" {
		t, err := time.Parse("2006-01-02 15:04:05", createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created_at: %w", err)
		}
		cart.CreatedAt = &t
	}
	fmt.Printf("DEBUG: Cart found: ID=%d, CustomerID=%d\n", cart.ShoppingCartId, cart.CustomerId)
	// Get items for this cart
	itemsQuery := `
		SELECT product_id, quantity
		FROM cart_items
		WHERE shopping_cart_id = ?
		ORDER BY created_at ASC
	`

	rows, err := DB.QueryContext(ctx, itemsQuery, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cart.Items = []api.CartItem{}
	// Iterate through results
	for rows.Next() {
		var item api.CartItem
		if err := rows.Scan(&item.ProductId, &item.Quantity); err != nil {
			return nil, err
		}
		cart.Items = append(cart.Items, item)
	}
	// Check for iteration errors
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cart items: %w", err)
	}

	return cart, nil
}

// AddItemToCart adds or updates an item in the cart (MySQL)
func (m *mysqlStore) AddItemToCart(ctx context.Context, cartID, productID, quantity int32) error {
	query := `
		INSERT INTO cart_items (shopping_cart_id, product_id, quantity)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE quantity = quantity + ?
	`

	_, err := DB.ExecContext(ctx, query, cartID, productID, quantity, quantity)
	if err != nil {
		return fmt.Errorf("failed to add item: %w", err)
	}

	return nil
}
