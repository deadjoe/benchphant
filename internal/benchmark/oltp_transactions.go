package benchmark

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"time"
)

// OLTPTransaction represents an OLTP transaction
type OLTPTransaction struct {
	Type    string
	Execute func(ctx context.Context, tx *sql.Tx) error
}

// NewOrderTransaction creates a new New-Order transaction
func NewOrderTransaction(warehouseID int) *OLTPTransaction {
	return &OLTPTransaction{
		Type: "new-order",
		Execute: func(ctx context.Context, tx *sql.Tx) error {
			// 1. Get customer info
			var customerID int
			err := tx.QueryRowContext(ctx,
				"SELECT c_id FROM customer WHERE c_w_id = ? ORDER BY RAND() LIMIT 1",
				warehouseID).Scan(&customerID)
			if err != nil {
				return fmt.Errorf("failed to get customer: %w", err)
			}

			// 2. Create new order
			orderID := rand.Int63()
			_, err = tx.ExecContext(ctx,
				"INSERT INTO orders (o_id, o_c_id, o_w_id, o_entry_d) VALUES (?, ?, ?, ?)",
				orderID, customerID, warehouseID, time.Now())
			if err != nil {
				return fmt.Errorf("failed to create order: %w", err)
			}

			// 3. Add order lines (1-5 items)
			numItems := rand.Intn(5) + 1
			for i := 0; i < numItems; i++ {
				itemID := rand.Int63n(100000) + 1
				quantity := rand.Int31n(10) + 1

				// Check stock and update
				var stockQuantity int32
				err = tx.QueryRowContext(ctx,
					"SELECT s_quantity FROM stock WHERE s_i_id = ? AND s_w_id = ? FOR UPDATE",
					itemID, warehouseID).Scan(&stockQuantity)
				if err != nil {
					return fmt.Errorf("failed to get stock: %w", err)
				}

				if stockQuantity < quantity {
					quantity = stockQuantity
				}

				// Update stock
				_, err = tx.ExecContext(ctx,
					"UPDATE stock SET s_quantity = s_quantity - ? WHERE s_i_id = ? AND s_w_id = ?",
					quantity, itemID, warehouseID)
				if err != nil {
					return fmt.Errorf("failed to update stock: %w", err)
				}

				// Add order line
				_, err = tx.ExecContext(ctx,
					"INSERT INTO order_line (ol_o_id, ol_number, ol_i_id, ol_quantity) VALUES (?, ?, ?, ?)",
					orderID, i+1, itemID, quantity)
				if err != nil {
					return fmt.Errorf("failed to add order line: %w", err)
				}
			}

			return nil
		},
	}
}

// PaymentTransaction creates a new Payment transaction
func PaymentTransaction(warehouseID int) *OLTPTransaction {
	return &OLTPTransaction{
		Type: "payment",
		Execute: func(ctx context.Context, tx *sql.Tx) error {
			// 1. Get customer
			var customerID int
			err := tx.QueryRowContext(ctx,
				"SELECT c_id FROM customer WHERE c_w_id = ? ORDER BY RAND() LIMIT 1",
				warehouseID).Scan(&customerID)
			if err != nil {
				return fmt.Errorf("failed to get customer: %w", err)
			}

			// 2. Update customer balance
			amount := rand.Float64()*5000 + 1
			_, err = tx.ExecContext(ctx,
				"UPDATE customer SET c_balance = c_balance - ? WHERE c_id = ? AND c_w_id = ?",
				amount, customerID, warehouseID)
			if err != nil {
				return fmt.Errorf("failed to update customer balance: %w", err)
			}

			// 3. Add payment history
			_, err = tx.ExecContext(ctx,
				"INSERT INTO history (h_c_id, h_w_id, h_amount, h_date) VALUES (?, ?, ?, ?)",
				customerID, warehouseID, amount, time.Now())
			if err != nil {
				return fmt.Errorf("failed to add payment history: %w", err)
			}

			return nil
		},
	}
}

// OrderStatusTransaction creates a new Order-Status transaction
func OrderStatusTransaction(warehouseID int) *OLTPTransaction {
	return &OLTPTransaction{
		Type: "order-status",
		Execute: func(ctx context.Context, tx *sql.Tx) error {
			// 1. Get customer's last order
			var orderID int64
			err := tx.QueryRowContext(ctx,
				`SELECT o.o_id 
				FROM orders o 
				JOIN customer c ON o.o_c_id = c.c_id 
				WHERE c.c_w_id = ? 
				ORDER BY o.o_entry_d DESC 
				LIMIT 1`,
				warehouseID).Scan(&orderID)
			if err != nil {
				return fmt.Errorf("failed to get last order: %w", err)
			}

			// 2. Get order lines
			rows, err := tx.QueryContext(ctx,
				"SELECT ol_i_id, ol_quantity FROM order_line WHERE ol_o_id = ?",
				orderID)
			if err != nil {
				return fmt.Errorf("failed to get order lines: %w", err)
			}
			defer rows.Close()

			for rows.Next() {
				var itemID int64
				var quantity int32
				if err := rows.Scan(&itemID, &quantity); err != nil {
					return fmt.Errorf("failed to scan order line: %w", err)
				}
				// In a real implementation, we might do something with this data
			}

			return rows.Err()
		},
	}
}

// DeliveryTransaction creates a new Delivery transaction
func DeliveryTransaction(warehouseID int) *OLTPTransaction {
	return &OLTPTransaction{
		Type: "delivery",
		Execute: func(ctx context.Context, tx *sql.Tx) error {
			// Process a batch of orders (1-10)
			numOrders := rand.Intn(10) + 1

			// Get undelivered orders
			rows, err := tx.QueryContext(ctx,
				`SELECT o_id, o_c_id 
				FROM orders 
				WHERE o_w_id = ? AND o_carrier_id IS NULL 
				ORDER BY o_entry_d 
				LIMIT ?`,
				warehouseID, numOrders)
			if err != nil {
				return fmt.Errorf("failed to get undelivered orders: %w", err)
			}
			defer rows.Close()

			for rows.Next() {
				var orderID int64
				var customerID int
				if err := rows.Scan(&orderID, &customerID); err != nil {
					return fmt.Errorf("failed to scan order: %w", err)
				}

				// Assign carrier
				carrierID := rand.Intn(10) + 1
				_, err = tx.ExecContext(ctx,
					"UPDATE orders SET o_carrier_id = ? WHERE o_id = ?",
					carrierID, orderID)
				if err != nil {
					return fmt.Errorf("failed to update order carrier: %w", err)
				}

				// Update delivery date for order lines
				_, err = tx.ExecContext(ctx,
					"UPDATE order_line SET ol_delivery_d = ? WHERE ol_o_id = ?",
					time.Now(), orderID)
				if err != nil {
					return fmt.Errorf("failed to update order line delivery date: %w", err)
				}
			}

			return rows.Err()
		},
	}
}

// StockLevelTransaction creates a new Stock-Level transaction
func StockLevelTransaction(warehouseID int) *OLTPTransaction {
	return &OLTPTransaction{
		Type: "stock-level",
		Execute: func(ctx context.Context, tx *sql.Tx) error {
			// Get items with stock below threshold
			threshold := rand.Int31n(10) + 1

			rows, err := tx.QueryContext(ctx,
				`SELECT s_i_id, s_quantity 
				FROM stock 
				WHERE s_w_id = ? AND s_quantity < ? 
				ORDER BY s_quantity`,
				warehouseID, threshold)
			if err != nil {
				return fmt.Errorf("failed to get low stock items: %w", err)
			}
			defer rows.Close()

			for rows.Next() {
				var itemID int64
				var quantity int32
				if err := rows.Scan(&itemID, &quantity); err != nil {
					return fmt.Errorf("failed to scan stock item: %w", err)
				}
				// In a real implementation, we might do something with this data
			}

			return rows.Err()
		},
	}
}
