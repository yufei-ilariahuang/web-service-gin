package database

import (
	"fmt"
)

// RunMigrations executes database migrations
func RunMigrations() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	// Create shopping_carts table
	createShoppingCartsTable := "CREATE TABLE IF NOT EXISTS shopping_carts (" +
		"id INT PRIMARY KEY AUTO_INCREMENT," +
		"customer_id INT NOT NULL," +
		"created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP," +
		"updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP," +
		"INDEX idx_customer_id (customer_id)," +
		"INDEX idx_created_at (created_at)" +
		") ENGINE=InnoDB"

	if _, err := DB.Exec(createShoppingCartsTable); err != nil {
		return fmt.Errorf("failed to create shopping_carts table: %w", err)
	}

	// Create cart_items table
	createCartItemsTable := "CREATE TABLE IF NOT EXISTS cart_items (" +
		"id INT PRIMARY KEY AUTO_INCREMENT," +
		"shopping_cart_id INT NOT NULL," +
		"product_id INT NOT NULL," +
		"quantity INT NOT NULL," +
		"created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP," +
		"UNIQUE KEY unique_cart_product (shopping_cart_id, product_id)," +
		"FOREIGN KEY fk_cart (shopping_cart_id) REFERENCES shopping_carts(id) ON DELETE CASCADE," +
		"INDEX idx_shopping_cart_id (shopping_cart_id)," +
		"INDEX idx_product_id (product_id)" +
		") ENGINE=InnoDB"

	if _, err := DB.Exec(createCartItemsTable); err != nil {
		return fmt.Errorf("failed to create cart_items table: %w", err)
	}

	return nil
}
