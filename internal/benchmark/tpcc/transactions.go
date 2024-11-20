package tpcc

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"time"
)

// Transaction represents a TPC-C transaction
type Transaction interface {
	Execute(ctx context.Context) error
}

// NewOrder implements the New-Order transaction
type NewOrder struct {
	tx       *sql.Tx
	db       *sql.DB
	wID      int
	dID      int
	cID      int
	itemIDs  []int
	supplyWs []int
	qtys     []int
	allLocal bool
}

// Execute runs the New-Order transaction
func (t *NewOrder) Execute(ctx context.Context) error {
	var err error
	t.tx, err = t.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer t.tx.Rollback()

	// Get warehouse tax rate
	var wTax float64
	err = t.tx.QueryRowContext(ctx,
		"SELECT w_tax FROM warehouse WHERE w_id = $1",
		t.wID,
	).Scan(&wTax)
	if err != nil {
		return err
	}

	// Get district information
	var dTax float64
	var dNextOID int
	err = t.tx.QueryRowContext(ctx,
		"SELECT d_tax, d_next_o_id FROM district WHERE d_w_id = $1 AND d_id = $2",
		t.wID, t.dID,
	).Scan(&dTax, &dNextOID)
	if err != nil {
		return err
	}

	// Update district next order ID
	_, err = t.tx.ExecContext(ctx,
		"UPDATE district SET d_next_o_id = d_next_o_id + 1 WHERE d_w_id = $1 AND d_id = $2",
		t.wID, t.dID,
	)
	if err != nil {
		return err
	}

	// Get customer information
	var cDiscount float64
	var cLast, cCredit string
	err = t.tx.QueryRowContext(ctx,
		"SELECT c_discount, c_last, c_credit FROM customer WHERE c_w_id = $1 AND c_d_id = $2 AND c_id = $3",
		t.wID, t.dID, t.cID,
	).Scan(&cDiscount, &cLast, &cCredit)
	if err != nil {
		return err
	}

	// Create new order
	_, err = t.tx.ExecContext(ctx,
		"INSERT INTO orders (o_id, o_d_id, o_w_id, o_c_id, o_entry_d, o_ol_cnt, o_all_local) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		dNextOID, t.dID, t.wID, t.cID, time.Now(), len(t.itemIDs), t.allLocal,
	)
	if err != nil {
		return err
	}

	// Create new order entry
	_, err = t.tx.ExecContext(ctx,
		"INSERT INTO new_order (no_o_id, no_d_id, no_w_id) VALUES ($1, $2, $3)",
		dNextOID, t.dID, t.wID,
	)
	if err != nil {
		return err
	}

	// Process each order line
	for i, itemID := range t.itemIDs {
		supplyWID := t.supplyWs[i]
		quantity := t.qtys[i]

		// Get item information
		var iPrice float64
		var iName string
		err = t.tx.QueryRowContext(ctx,
			"SELECT i_price, i_name FROM item WHERE i_id = $1",
			itemID,
		).Scan(&iPrice, &iName)
		if err != nil {
			return err
		}

		// Get stock information and update
		var sQuantity int
		var sData, sDist string
		err = t.tx.QueryRowContext(ctx,
			fmt.Sprintf("SELECT s_quantity, s_data, s_dist_%02d FROM stock WHERE s_i_id = $1 AND s_w_id = $2",
				t.dID),
			itemID, supplyWID,
		).Scan(&sQuantity, &sData, &sDist)
		if err != nil {
			return err
		}

		// Update stock
		newQuantity := sQuantity - quantity
		if newQuantity < 10 {
			newQuantity += 91
		}

		_, err = t.tx.ExecContext(ctx,
			`UPDATE stock 
			SET s_quantity = $1,
				s_ytd = s_ytd + $2,
				s_order_cnt = s_order_cnt + 1,
				s_remote_cnt = s_remote_cnt + $3
			WHERE s_i_id = $4 AND s_w_id = $5`,
			newQuantity,
			quantity,
			boolToInt(supplyWID != t.wID),
			itemID,
			supplyWID,
		)
		if err != nil {
			return err
		}

		// Create order line
		amount := float64(quantity) * iPrice
		_, err = t.tx.ExecContext(ctx,
			`INSERT INTO order_line (
				ol_o_id, ol_d_id, ol_w_id, ol_number,
				ol_i_id, ol_supply_w_id, ol_quantity,
				ol_amount, ol_dist_info
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			dNextOID,
			t.dID,
			t.wID,
			i+1,
			itemID,
			supplyWID,
			quantity,
			amount,
			sDist,
		)
		if err != nil {
			return err
		}
	}

	return t.tx.Commit()
}

// Payment implements the Payment transaction
type Payment struct {
	tx     *sql.Tx
	db     *sql.DB
	wID    int
	dID    int
	cID    int
	amount float64
}

// Execute runs the Payment transaction
func (t *Payment) Execute(ctx context.Context) error {
	var err error
	t.tx, err = t.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer t.tx.Rollback()

	// Update warehouse YTD
	_, err = t.tx.ExecContext(ctx,
		"UPDATE warehouse SET w_ytd = w_ytd + $1 WHERE w_id = $2",
		t.amount, t.wID,
	)
	if err != nil {
		return err
	}

	// Update district YTD
	_, err = t.tx.ExecContext(ctx,
		"UPDATE district SET d_ytd = d_ytd + $1 WHERE d_w_id = $2 AND d_id = $3",
		t.amount, t.wID, t.dID,
	)
	if err != nil {
		return err
	}

	// Update customer
	_, err = t.tx.ExecContext(ctx,
		`UPDATE customer 
		SET c_balance = c_balance - $1,
			c_ytd_payment = c_ytd_payment + $1,
			c_payment_cnt = c_payment_cnt + 1
		WHERE c_w_id = $2 AND c_d_id = $3 AND c_id = $4`,
		t.amount, t.wID, t.dID, t.cID,
	)
	if err != nil {
		return err
	}

	// Insert history record
	_, err = t.tx.ExecContext(ctx,
		`INSERT INTO history (
			h_c_id, h_c_d_id, h_c_w_id,
			h_d_id, h_w_id,
			h_date, h_amount, h_data
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		t.cID, t.dID, t.wID,
		t.dID, t.wID,
		time.Now(), t.amount,
		fmt.Sprintf("W%dD%d", t.wID, t.dID),
	)
	if err != nil {
		return err
	}

	return t.tx.Commit()
}

// OrderStatus implements the Order-Status transaction
type OrderStatus struct {
	tx  *sql.Tx
	db  *sql.DB
	wID int
	dID int
	cID int
}

// Execute runs the Order-Status transaction
func (t *OrderStatus) Execute(ctx context.Context) error {
	var err error
	t.tx, err = t.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer t.tx.Rollback()

	// Get customer's last order
	var oID int
	var oEntryD time.Time
	var oCarrierID sql.NullInt64
	err = t.tx.QueryRowContext(ctx,
		`SELECT o_id, o_entry_d, o_carrier_id
		FROM orders
		WHERE o_w_id = $1 AND o_d_id = $2 AND o_c_id = $3
		ORDER BY o_id DESC LIMIT 1`,
		t.wID, t.dID, t.cID,
	).Scan(&oID, &oEntryD, &oCarrierID)
	if err != nil {
		return err
	}

	// Get order lines
	rows, err := t.tx.QueryContext(ctx,
		`SELECT ol_i_id, ol_supply_w_id, ol_quantity, ol_amount, ol_delivery_d
		FROM order_line
		WHERE ol_w_id = $1 AND ol_d_id = $2 AND ol_o_id = $3`,
		t.wID, t.dID, oID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var olIID, olSupplyWID int
		var olQuantity int
		var olAmount float64
		var olDeliveryD sql.NullTime
		err = rows.Scan(&olIID, &olSupplyWID, &olQuantity, &olAmount, &olDeliveryD)
		if err != nil {
			return err
		}
		// Process order line information as needed
	}

	return t.tx.Commit()
}

// Delivery implements the Delivery transaction
type Delivery struct {
	tx        *sql.Tx
	db        *sql.DB
	wID       int
	carrierID int
}

// Execute runs the Delivery transaction
func (t *Delivery) Execute(ctx context.Context) error {
	var err error
	t.tx, err = t.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer t.tx.Rollback()

	// Process each district
	for dID := 1; dID <= 10; dID++ {
		// Get oldest undelivered order
		var oID int
		var oCID int
		err = t.tx.QueryRowContext(ctx,
			`SELECT no_o_id, o_c_id
			FROM new_order
			JOIN orders ON o_id = no_o_id AND o_w_id = no_w_id AND o_d_id = no_d_id
			WHERE no_w_id = $1 AND no_d_id = $2
			ORDER BY no_o_id ASC
			LIMIT 1`,
			t.wID, dID,
		).Scan(&oID, &oCID)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return err
		}

		// Delete from new_order
		_, err = t.tx.ExecContext(ctx,
			"DELETE FROM new_order WHERE no_w_id = $1 AND no_d_id = $2 AND no_o_id = $3",
			t.wID, dID, oID,
		)
		if err != nil {
			return err
		}

		// Update order
		_, err = t.tx.ExecContext(ctx,
			"UPDATE orders SET o_carrier_id = $1 WHERE o_w_id = $2 AND o_d_id = $3 AND o_id = $4",
			t.carrierID, t.wID, dID, oID,
		)
		if err != nil {
			return err
		}

		// Update order lines
		var olTotal float64
		err = t.tx.QueryRowContext(ctx,
			`UPDATE order_line 
			SET ol_delivery_d = $1
			WHERE ol_w_id = $2 AND ol_d_id = $3 AND ol_o_id = $4
			RETURNING SUM(ol_amount)`,
			time.Now(), t.wID, dID, oID,
		).Scan(&olTotal)
		if err != nil {
			return err
		}

		// Update customer
		_, err = t.tx.ExecContext(ctx,
			`UPDATE customer
			SET c_balance = c_balance + $1,
				c_delivery_cnt = c_delivery_cnt + 1
			WHERE c_w_id = $2 AND c_d_id = $3 AND c_id = $4`,
			olTotal, t.wID, dID, oCID,
		)
		if err != nil {
			return err
		}
	}

	return t.tx.Commit()
}

// StockLevel implements the Stock-Level transaction
type StockLevel struct {
	tx        *sql.Tx
	db        *sql.DB
	wID       int
	dID       int
	threshold int
}

// Execute runs the Stock-Level transaction
func (t *StockLevel) Execute(ctx context.Context) error {
	var err error
	t.tx, err = t.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer t.tx.Rollback()

	// Get next order ID
	var dNextOID int
	err = t.tx.QueryRowContext(ctx,
		"SELECT d_next_o_id FROM district WHERE d_w_id = $1 AND d_id = $2",
		t.wID, t.dID,
	).Scan(&dNextOID)
	if err != nil {
		return err
	}

	// Count items below threshold
	var lowStock int
	err = t.tx.QueryRowContext(ctx,
		`SELECT COUNT(DISTINCT(s_i_id))
		FROM order_line, stock
		WHERE ol_w_id = $1
		AND ol_d_id = $2
		AND ol_o_id < $3
		AND ol_o_id >= $4
		AND s_w_id = $1
		AND s_i_id = ol_i_id
		AND s_quantity < $5`,
		t.wID, t.dID,
		dNextOID,
		dNextOID-20,
		t.threshold,
	).Scan(&lowStock)
	if err != nil {
		return err
	}

	return t.tx.Commit()
}

// Helper function to convert bool to int
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
