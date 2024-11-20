package tpcc

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// TransactionExecutor manages and executes TPC-C transactions
type TransactionExecutor struct {
	db     *sql.DB
	config *Config
	mu     sync.Mutex
}

// NewTransactionExecutor creates a new transaction executor
func NewTransactionExecutor(db *sql.DB, config *Config) *TransactionExecutor {
	return &TransactionExecutor{
		db:     db,
		config: config,
	}
}

// ExecuteNewOrder executes a New-Order transaction
func (e *TransactionExecutor) ExecuteNewOrder(ctx context.Context, tx *NewOrder) error {
	// Start transaction
	dbTx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer dbTx.Rollback()

	// Get warehouse tax rate
	var wTax float64
	err = dbTx.QueryRowContext(ctx,
		"SELECT w_tax FROM warehouse WHERE w_id = ?",
		tx.wID).Scan(&wTax)
	if err != nil {
		return fmt.Errorf("get warehouse tax: %w", err)
	}

	// Get district tax rate and next order ID
	var dTax float64
	var dNextOID int
	err = dbTx.QueryRowContext(ctx,
		"SELECT d_tax, d_next_o_id FROM district WHERE d_w_id = ? AND d_id = ?",
		tx.wID, tx.dID).Scan(&dTax, &dNextOID)
	if err != nil {
		return fmt.Errorf("get district info: %w", err)
	}

	// Update district next order ID
	_, err = dbTx.ExecContext(ctx,
		"UPDATE district SET d_next_o_id = ? WHERE d_w_id = ? AND d_id = ?",
		dNextOID+1, tx.wID, tx.dID)
	if err != nil {
		return fmt.Errorf("update district next order ID: %w", err)
	}

	// Get customer discount rate
	var cDiscount float64
	var cLast, cCredit string
	err = dbTx.QueryRowContext(ctx,
		"SELECT c_discount, c_last, c_credit FROM customer WHERE c_w_id = ? AND c_d_id = ? AND c_id = ?",
		tx.wID, tx.dID, tx.cID).Scan(&cDiscount, &cLast, &cCredit)
	if err != nil {
		return fmt.Errorf("get customer info: %w", err)
	}

	// Create new order
	now := time.Now()
	_, err = dbTx.ExecContext(ctx,
		"INSERT INTO orders (o_id, o_d_id, o_w_id, o_c_id, o_entry_d, o_ol_cnt, o_all_local) VALUES (?, ?, ?, ?, ?, ?, ?)",
		dNextOID, tx.dID, tx.wID, tx.cID, now, len(tx.itemIDs), tx.allLocal)
	if err != nil {
		return fmt.Errorf("create order: %w", err)
	}

	// Create new order entry
	_, err = dbTx.ExecContext(ctx,
		"INSERT INTO new_order (no_o_id, no_d_id, no_w_id) VALUES (?, ?, ?)",
		dNextOID, tx.dID, tx.wID)
	if err != nil {
		return fmt.Errorf("create new order entry: %w", err)
	}

	// Process order lines
	var totalAmount float64
	for i, itemID := range tx.itemIDs {
		// Get item price and name
		var iPrice float64
		var iName string
		err = dbTx.QueryRowContext(ctx,
			"SELECT i_price, i_name FROM item WHERE i_id = ?",
			itemID).Scan(&iPrice, &iName)
		if err != nil {
			return fmt.Errorf("get item info: %w", err)
		}

		// Get stock info and update
		var sQuantity int
		var sDistInfo string
		var sYtd int
		var sOrderCnt int
		var sRemoteCnt int

		err = dbTx.QueryRowContext(ctx,
			"SELECT s_quantity, s_dist_01, s_ytd, s_order_cnt, s_remote_cnt FROM stock WHERE s_i_id = ? AND s_w_id = ?",
			itemID, tx.supplyWs[i]).Scan(&sQuantity, &sDistInfo, &sYtd, &sOrderCnt, &sRemoteCnt)
		if err != nil {
			return fmt.Errorf("get stock info: %w", err)
		}

		// Update stock
		newQuantity := sQuantity - tx.qtys[i]
		if newQuantity < 10 {
			newQuantity += 91
		}

		newRemoteCnt := sRemoteCnt
		if tx.supplyWs[i] != tx.wID {
			newRemoteCnt++
		}

		_, err = dbTx.ExecContext(ctx,
			"UPDATE stock SET s_quantity = ?, s_ytd = ?, s_order_cnt = ?, s_remote_cnt = ? WHERE s_i_id = ? AND s_w_id = ?",
			newQuantity, sYtd+tx.qtys[i], sOrderCnt+1, newRemoteCnt, itemID, tx.supplyWs[i])
		if err != nil {
			return fmt.Errorf("update stock: %w", err)
		}

		// Calculate amount
		amount := float64(tx.qtys[i]) * iPrice
		totalAmount += amount

		// Create order line
		_, err = dbTx.ExecContext(ctx,
			"INSERT INTO order_line (ol_o_id, ol_d_id, ol_w_id, ol_number, ol_i_id, ol_supply_w_id, ol_quantity, ol_amount, ol_dist_info) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
			dNextOID, tx.dID, tx.wID, i+1, itemID, tx.supplyWs[i], tx.qtys[i], amount, sDistInfo)
		if err != nil {
			return fmt.Errorf("create order line: %w", err)
		}
	}

	// Calculate total amount with tax
	totalAmount = totalAmount * (1 + wTax + dTax) * (1 - cDiscount)

	// Commit transaction
	if err = dbTx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// ExecutePayment executes a Payment transaction
func (e *TransactionExecutor) ExecutePayment(ctx context.Context, tx *Payment) error {
	// Start transaction
	dbTx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer dbTx.Rollback()

	// Update warehouse
	_, err = dbTx.ExecContext(ctx,
		"UPDATE warehouse SET w_ytd = w_ytd + ? WHERE w_id = ?",
		tx.amount, tx.wID)
	if err != nil {
		return fmt.Errorf("update warehouse: %w", err)
	}

	// Update district
	_, err = dbTx.ExecContext(ctx,
		"UPDATE district SET d_ytd = d_ytd + ? WHERE d_w_id = ? AND d_id = ?",
		tx.amount, tx.wID, tx.dID)
	if err != nil {
		return fmt.Errorf("update district: %w", err)
	}

	// Update customer
	_, err = dbTx.ExecContext(ctx,
		"UPDATE customer SET c_balance = c_balance - ?, c_ytd_payment = c_ytd_payment + ?, c_payment_cnt = c_payment_cnt + 1 WHERE c_w_id = ? AND c_d_id = ? AND c_id = ?",
		tx.amount, tx.amount, tx.wID, tx.dID, tx.cID)
	if err != nil {
		return fmt.Errorf("update customer: %w", err)
	}

	// Commit transaction
	if err = dbTx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// ExecuteOrderStatus executes an Order-Status transaction
func (e *TransactionExecutor) ExecuteOrderStatus(ctx context.Context, tx *OrderStatus) error {
	// Get customer's last order
	var lastOrderID int
	err := e.db.QueryRowContext(ctx,
		"SELECT o_id FROM orders WHERE o_w_id = ? AND o_d_id = ? AND o_c_id = ? ORDER BY o_id DESC LIMIT 1",
		tx.wID, tx.dID, tx.cID).Scan(&lastOrderID)
	if err != nil {
		return fmt.Errorf("get last order: %w", err)
	}

	// Get order lines
	rows, err := e.db.QueryContext(ctx,
		"SELECT ol_i_id, ol_supply_w_id, ol_quantity, ol_amount, ol_delivery_d FROM order_line WHERE ol_w_id = ? AND ol_d_id = ? AND ol_o_id = ?",
		tx.wID, tx.dID, lastOrderID)
	if err != nil {
		return fmt.Errorf("get order lines: %w", err)
	}
	defer rows.Close()

	return nil
}

// ExecuteDelivery executes a Delivery transaction
func (e *TransactionExecutor) ExecuteDelivery(ctx context.Context, tx *Delivery) error {
	// Start transaction
	dbTx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer dbTx.Rollback()

	now := time.Now()

	// Process each district
	for dID := 1; dID <= 10; dID++ {
		// Get oldest new order
		var oID int
		err := dbTx.QueryRowContext(ctx,
			"SELECT no_o_id FROM new_order WHERE no_w_id = ? AND no_d_id = ? ORDER BY no_o_id ASC LIMIT 1",
			tx.wID, dID).Scan(&oID)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return fmt.Errorf("get oldest new order: %w", err)
		}

		// Delete the new order
		_, err = dbTx.ExecContext(ctx,
			"DELETE FROM new_order WHERE no_w_id = ? AND no_d_id = ? AND no_o_id = ?",
			tx.wID, dID, oID)
		if err != nil {
			return fmt.Errorf("delete new order: %w", err)
		}

		// Get customer ID and total amount
		var cID int
		var totalAmount float64
		err = dbTx.QueryRowContext(ctx,
			"SELECT o_c_id, SUM(ol_amount) FROM orders JOIN order_line ON ol_w_id = o_w_id AND ol_d_id = o_d_id AND ol_o_id = o_id WHERE o_w_id = ? AND o_d_id = ? AND o_id = ? GROUP BY o_c_id",
			tx.wID, dID, oID).Scan(&cID, &totalAmount)
		if err != nil {
			return fmt.Errorf("get order info: %w", err)
		}

		// Update order
		_, err = dbTx.ExecContext(ctx,
			"UPDATE orders SET o_carrier_id = ? WHERE o_w_id = ? AND o_d_id = ? AND o_id = ?",
			tx.carrierID, tx.wID, dID, oID)
		if err != nil {
			return fmt.Errorf("update order: %w", err)
		}

		// Update order lines
		_, err = dbTx.ExecContext(ctx,
			"UPDATE order_line SET ol_delivery_d = ? WHERE ol_w_id = ? AND ol_d_id = ? AND ol_o_id = ?",
			now, tx.wID, dID, oID)
		if err != nil {
			return fmt.Errorf("update order lines: %w", err)
		}

		// Update customer
		_, err = dbTx.ExecContext(ctx,
			"UPDATE customer SET c_balance = c_balance + ?, c_delivery_cnt = c_delivery_cnt + 1 WHERE c_w_id = ? AND c_d_id = ? AND c_id = ?",
			totalAmount, tx.wID, dID, cID)
		if err != nil {
			return fmt.Errorf("update customer: %w", err)
		}
	}

	// Commit transaction
	if err = dbTx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// ExecuteStockLevel executes a Stock-Level transaction
func (e *TransactionExecutor) ExecuteStockLevel(ctx context.Context, tx *StockLevel) error {
	// Get district's last 20 orders
	var lowStockCount int
	err := e.db.QueryRowContext(ctx,
		`SELECT COUNT(DISTINCT(s_i_id)) 
		FROM stock 
		JOIN order_line ON ol_i_id = s_i_id
		WHERE s_w_id = ? 
		AND ol_w_id = ? 
		AND ol_d_id = ? 
		AND ol_o_id >= (
			SELECT d_next_o_id - 20 
			FROM district 
			WHERE d_w_id = ? 
			AND d_id = ?
		)
		AND s_quantity < ?`,
		tx.wID, tx.wID, tx.dID, tx.wID, tx.dID, tx.threshold).Scan(&lowStockCount)
	if err != nil {
		return fmt.Errorf("get low stock count: %w", err)
	}

	return nil
}
