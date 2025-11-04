package database

import (
    "context"
    "fmt"
    "src/api"
)

// Store is the storage abstraction for shopping cart operations.
// Implementations: MySQL (existing) and DynamoDB (new).
type Store interface {
    CreateCart(ctx context.Context, customerID int32) (int32, error)
    GetCart(ctx context.Context, cartID int32) (*api.ShoppingCart, error)
    AddItemToCart(ctx context.Context, cartID, productID, quantity int32) error
}

// Impl is the active store implementation. Must be set during startup.
var Impl Store

// SetStore selects the active store implementation.
func SetStore(s Store) {
    Impl = s
}

// CreateCart delegates to the configured store implementation.
func CreateCart(ctx context.Context, customerID int32) (int32, error) {
    if Impl == nil {
        return 0, fmt.Errorf("store not configured")
    }
    return Impl.CreateCart(ctx, customerID)
}

// GetCart delegates to the configured store implementation.
func GetCart(ctx context.Context, cartID int32) (*api.ShoppingCart, error) {
    if Impl == nil {
        return nil, fmt.Errorf("store not configured")
    }
    return Impl.GetCart(ctx, cartID)
}

// AddItemToCart delegates to the configured store implementation.
func AddItemToCart(ctx context.Context, cartID, productID, quantity int32) error {
    if Impl == nil {
        return fmt.Errorf("store not configured")
    }
    return Impl.AddItemToCart(ctx, cartID, productID, quantity)
}
