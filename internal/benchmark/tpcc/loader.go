package tpcc

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Loader handles the generation and loading of TPC-C test data
type Loader struct {
	db     *sql.DB
	config *Config
	wg     sync.WaitGroup
}

// NewLoader creates a new data loader
func NewLoader(db *sql.DB, config *Config) *Loader {
	return &Loader{
		db:     db,
		config: config,
	}
}

// Load generates and loads all TPC-C test data
func (l *Loader) Load(ctx context.Context) error {
	// Load items first (they are referenced by other tables)
	if err := l.loadItems(ctx); err != nil {
		return fmt.Errorf("failed to load items: %w", err)
	}

	// Load warehouses and related data in parallel
	for w := 1; w <= l.config.Warehouses; w++ {
		w := w // Capture for goroutine
		l.wg.Add(1)
		go func() {
			defer l.wg.Done()
			if err := l.loadWarehouse(ctx, w); err != nil {
				fmt.Printf("Error loading warehouse %d: %v\n", w, err)
			}
		}()
	}

	l.wg.Wait()
	return nil
}

// loadItems loads the item table
func (l *Loader) loadItems(ctx context.Context) error {
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO item (i_id, i_im_id, i_name, i_price, i_data)
		VALUES ($1, $2, $3, $4, $5)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Load 100,000 items as per TPC-C spec
	for i := 1; i <= 100000; i++ {
		if err := l.insertItem(ctx, stmt, i); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// loadWarehouse loads a warehouse and all its related data
func (l *Loader) loadWarehouse(ctx context.Context, wID int) error {
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert warehouse
	if err := l.insertWarehouse(ctx, tx, wID); err != nil {
		return err
	}

	// Load stock for this warehouse
	if err := l.loadStock(ctx, tx, wID); err != nil {
		return err
	}

	// Load districts for this warehouse
	for d := 1; d <= 10; d++ {
		if err := l.loadDistrict(ctx, tx, wID, d); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// insertItem inserts a single item
func (l *Loader) insertItem(ctx context.Context, stmt *sql.Stmt, id int) error {
	name := fmt.Sprintf("Item-%d", id)
	price := float64(rand.Intn(9900)+100) / 100.0 // Price between 1.00 and 100.00
	data := randomString(26, 50)

	_, err := stmt.ExecContext(ctx, id, rand.Intn(10000), name, price, data)
	return err
}

// insertWarehouse inserts a single warehouse
func (l *Loader) insertWarehouse(ctx context.Context, tx *sql.Tx, id int) error {
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO warehouse (w_id, w_name, w_street_1, w_street_2, w_city, w_state, w_zip, w_tax, w_ytd)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	name := fmt.Sprintf("Warehouse-%d", id)
	tax := float64(rand.Intn(2000)) / 10000.0 // Tax between 0 and 0.2000

	_, err = stmt.ExecContext(ctx,
		id,
		name,
		randomString(10, 20),
		randomString(10, 20),
		randomString(10, 20),
		randomState(),
		randomZIP(),
		tax,
		300000.00, // Initial YTD
	)
	return err
}

// loadStock loads stock for all items in a warehouse
func (l *Loader) loadStock(ctx context.Context, tx *sql.Tx, wID int) error {
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO stock (
			s_i_id, s_w_id, s_quantity,
			s_dist_01, s_dist_02, s_dist_03, s_dist_04, s_dist_05,
			s_dist_06, s_dist_07, s_dist_08, s_dist_09, s_dist_10,
			s_ytd, s_order_cnt, s_remote_cnt, s_data
		) VALUES (
			$1, $2, $3,
			$4, $5, $6, $7, $8, $9, $10, $11, $12, $13,
			$14, $15, $16, $17
		)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := 1; i <= 100000; i++ {
		quantity := rand.Intn(91) + 10 // Stock quantity between 10 and 100
		data := randomString(26, 50)

		_, err = stmt.ExecContext(ctx,
			i, wID, quantity,
			randomString(24, 24), // s_dist_01
			randomString(24, 24), // s_dist_02
			randomString(24, 24), // s_dist_03
			randomString(24, 24), // s_dist_04
			randomString(24, 24), // s_dist_05
			randomString(24, 24), // s_dist_06
			randomString(24, 24), // s_dist_07
			randomString(24, 24), // s_dist_08
			randomString(24, 24), // s_dist_09
			randomString(24, 24), // s_dist_10
			0,                    // s_ytd
			0,                    // s_order_cnt
			0,                    // s_remote_cnt
			data,                 // s_data
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// loadDistrict loads a district and its customers
func (l *Loader) loadDistrict(ctx context.Context, tx *sql.Tx, wID, dID int) error {
	// Insert district
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO district (
			d_id, d_w_id, d_name, d_street_1, d_street_2,
			d_city, d_state, d_zip, d_tax, d_ytd, d_next_o_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	name := fmt.Sprintf("District-%d", dID)
	tax := float64(rand.Intn(2000)) / 10000.0 // Tax between 0 and 0.2000

	_, err = stmt.ExecContext(ctx,
		dID,
		wID,
		name,
		randomString(10, 20),
		randomString(10, 20),
		randomString(10, 20),
		randomState(),
		randomZIP(),
		tax,
		30000.00, // Initial YTD
		3001,     // Next order ID
	)
	if err != nil {
		return err
	}

	// Load customers for this district
	return l.loadCustomers(ctx, tx, wID, dID)
}

// loadCustomers loads all customers for a district
func (l *Loader) loadCustomers(ctx context.Context, tx *sql.Tx, wID, dID int) error {
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO customer (
			c_id, c_d_id, c_w_id, c_first, c_middle, c_last,
			c_street_1, c_street_2, c_city, c_state, c_zip,
			c_phone, c_since, c_credit, c_credit_lim,
			c_discount, c_balance, c_ytd_payment,
			c_payment_cnt, c_delivery_cnt, c_data
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,
			$12, $13, $14, $15, $16, $17, $18, $19, $20, $21
		)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Load 3000 customers per district as per TPC-C spec
	for c := 1; c <= 3000; c++ {
		lastName := randomLastName(c)
		discount := float64(rand.Intn(5000)) / 10000.0 // Discount between 0 and 0.5000

		_, err = stmt.ExecContext(ctx,
			c,
			dID,
			wID,
			randomString(8, 16), // First name
			"OE",                // Middle name
			lastName,
			randomString(10, 20), // Street 1
			randomString(10, 20), // Street 2
			randomString(10, 20), // City
			randomState(),
			randomZIP(),
			randomString(16, 16), // Phone
			time.Now(),           // Since
			randomCredit(),
			50000.00, // Credit limit
			discount,
			-10.00,                 // Balance
			10.00,                  // YTD payment
			1,                      // Payment count
			0,                      // Delivery count
			randomString(300, 500), // Data
		)
		if err != nil {
			return err
		}

		// Create history record for this customer
		if err := l.insertHistory(ctx, tx, c, dID, wID); err != nil {
			return err
		}
	}

	return nil
}

// insertHistory inserts a history record for a customer
func (l *Loader) insertHistory(ctx context.Context, tx *sql.Tx, cID, dID, wID int) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO history (
			h_c_id, h_c_d_id, h_c_w_id,
			h_d_id, h_w_id, h_date,
			h_amount, h_data
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		cID,
		dID,
		wID,
		dID,
		wID,
		time.Now(),
		10.00,
		randomString(12, 24),
	)
	return err
}

// Helper functions for generating random data
func randomString(min, max int) string {
	length := min
	if max > min {
		length += rand.Intn(max - min + 1)
	}
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func randomState() string {
	states := []string{"AL", "AK", "AZ", "AR", "CA", "CO", "CT", "DE", "FL", "GA",
		"HI", "ID", "IL", "IN", "IA", "KS", "KY", "LA", "ME", "MD",
		"MA", "MI", "MN", "MS", "MO", "MT", "NE", "NV", "NH", "NJ",
		"NM", "NY", "NC", "ND", "OH", "OK", "OR", "PA", "RI", "SC",
		"SD", "TN", "TX", "UT", "VT", "VA", "WA", "WV", "WI", "WY"}
	return states[rand.Intn(len(states))]
}

func randomZIP() string {
	return fmt.Sprintf("%04d11111", rand.Intn(10000))
}

func randomCredit() string {
	if rand.Float64() < 0.1 {
		return "BC" // Bad credit (10% probability)
	}
	return "GC" // Good credit (90% probability)
}

func randomLastName(num int) string {
	// TPC-C spec requires a special distribution for customer last names
	syllables := []string{
		"BAR", "OUGHT", "ABLE", "PRI", "PRES",
		"ESE", "ANTI", "CALLY", "ATION", "EING",
	}

	n := ((num * 2147483647) % 10000) / 100
	i1 := n / 100
	i2 := (n % 100) / 10
	i3 := n % 10

	return syllables[i1] + syllables[i2] + syllables[i3]
}
