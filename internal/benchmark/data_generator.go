package benchmark

import (
	"database/sql"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// DataGenerator generates test data according to TPC-C specification
type DataGenerator struct {
	db             *sql.DB
	warehouseCount int
}

// NewDataGenerator creates a new data generator
func NewDataGenerator(db *sql.DB, warehouseCount int) *DataGenerator {
	return &DataGenerator{
		db:             db,
		warehouseCount: warehouseCount,
	}
}

// GenerateWarehouseData generates warehouse data
func (g *DataGenerator) GenerateWarehouseData() error {
	query := `INSERT INTO warehouse (w_id, w_name, w_street_1, w_street_2, w_city, w_state, w_zip, w_tax, w_ytd)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	for w := 1; w <= g.warehouseCount; w++ {
		_, err := g.db.Exec(query,
			w,                     // w_id
			fmt.Sprintf("W_%d", w), // w_name
			"Street 1",            // w_street_1
			"Street 2",            // w_street_2
			"City",                // w_city
			"ST",                  // w_state
			g.generateZip(),       // w_zip
			rand.Float64() * 0.2,  // w_tax
			300000.00,            // w_ytd
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// GenerateDistrictData generates district data
func (g *DataGenerator) GenerateDistrictData() error {
	query := `INSERT INTO district (d_id, d_w_id, d_name, d_street_1, d_street_2, d_city, d_state, d_zip, d_tax, d_ytd, d_next_o_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	for w := 1; w <= g.warehouseCount; w++ {
		for d := 1; d <= 10; d++ {
			_, err := g.db.Exec(query,
				d,                        // d_id
				w,                        // d_w_id
				fmt.Sprintf("D_%d_%d", w, d), // d_name
				"Street 1",               // d_street_1
				"Street 2",               // d_street_2
				"City",                   // d_city
				"ST",                     // d_state
				g.generateZip(),          // d_zip
				rand.Float64() * 0.2,     // d_tax
				30000.00,                // d_ytd
				3001,                    // d_next_o_id
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// GenerateCustomerData generates customer and history data
func (g *DataGenerator) GenerateCustomerData() error {
	customerQuery := `INSERT INTO customer (c_id, c_d_id, c_w_id, c_first, c_middle, c_last, c_street_1, c_street_2,
		c_city, c_state, c_zip, c_phone, c_since, c_credit, c_credit_lim, c_discount, c_balance, c_ytd_payment,
		c_payment_cnt, c_delivery_cnt, c_data)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	historyQuery := `INSERT INTO history (h_c_id, h_c_d_id, h_c_w_id, h_d_id, h_w_id, h_date, h_amount, h_data)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	for w := 1; w <= g.warehouseCount; w++ {
		for d := 1; d <= 10; d++ {
			for c := 1; c <= 3000; c++ {
				// Insert customer
				_, err := g.db.Exec(customerQuery,
					c,                    // c_id
					d,                    // c_d_id
					w,                    // c_w_id
					"FirstName",          // c_first
					"MI",                 // c_middle
					"LastName",           // c_last
					"Street 1",           // c_street_1
					"Street 2",           // c_street_2
					"City",               // c_city
					"ST",                 // c_state
					g.generateZip(),      // c_zip
					g.generatePhoneNumber(), // c_phone
					time.Now(),           // c_since
					"GC",                 // c_credit
					50000.00,             // c_credit_lim
					rand.Float64() * 0.5,  // c_discount
					-10.00,               // c_balance
					10.00,                // c_ytd_payment
					1,                    // c_payment_cnt
					0,                    // c_delivery_cnt
					"Customer Data",      // c_data
				)
				if err != nil {
					return err
				}

				// Insert history
				_, err = g.db.Exec(historyQuery,
					c,           // h_c_id
					d,           // h_c_d_id
					w,           // h_c_w_id
					d,           // h_d_id
					w,           // h_w_id
					time.Now(),  // h_date
					10.00,       // h_amount
					"History",   // h_data
				)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// GenerateItemData generates item data
func (g *DataGenerator) GenerateItemData() error {
	query := `INSERT INTO item (i_id, i_im_id, i_name, i_price, i_data)
		VALUES (?, ?, ?, ?, ?)`

	for i := 1; i <= 100000; i++ {
		_, err := g.db.Exec(query,
			i,                     // i_id
			i,                     // i_im_id
			fmt.Sprintf("Item_%d", i), // i_name
			rand.Float64() * 100.0,    // i_price
			"Item Data",           // i_data
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// GenerateStockData generates stock data
func (g *DataGenerator) GenerateStockData() error {
	query := `INSERT INTO stock (s_i_id, s_w_id, s_quantity, s_dist_01, s_dist_02, s_dist_03, s_dist_04,
		s_dist_05, s_dist_06, s_dist_07, s_dist_08, s_dist_09, s_dist_10, s_ytd, s_order_cnt, s_remote_cnt, s_data)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	for w := 1; w <= g.warehouseCount; w++ {
		for i := 1; i <= 100000; i++ {
			_, err := g.db.Exec(query,
				i,                // s_i_id
				w,                // s_w_id
				rand.Intn(91) + 10, // s_quantity (10-100)
				"dist_01",         // s_dist_01
				"dist_02",         // s_dist_02
				"dist_03",         // s_dist_03
				"dist_04",         // s_dist_04
				"dist_05",         // s_dist_05
				"dist_06",         // s_dist_06
				"dist_07",         // s_dist_07
				"dist_08",         // s_dist_08
				"dist_09",         // s_dist_09
				"dist_10",         // s_dist_10
				0,                // s_ytd
				0,                // s_order_cnt
				0,                // s_remote_cnt
				"Stock Data",      // s_data
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// GenerateOrderData generates order, order-line and new-order data
func (g *DataGenerator) GenerateOrderData() error {
	// 使用相同的随机数种子
	rnd := rand.New(rand.NewSource(42))

	orderQuery := `INSERT INTO orders (o_id, o_d_id, o_w_id, o_c_id, o_entry_d, o_carrier_id, o_ol_cnt, o_all_local)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	orderLineQuery := `INSERT INTO order_line (ol_o_id, ol_d_id, ol_w_id, ol_number, ol_i_id, ol_supply_w_id,
		ol_delivery_d, ol_quantity, ol_amount, ol_dist_info)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	newOrderQuery := `INSERT INTO new_order (no_o_id, no_d_id, no_w_id)
		VALUES (?, ?, ?)`

	// 在测试模式下只生成少量数据
	maxDistricts := 10
	maxOrders := 3000
	if testing.Testing() {
		maxDistricts = 1
		maxOrders = 10
	}

	for w := 1; w <= g.warehouseCount; w++ {
		for d := 1; d <= maxDistricts; d++ {
			for o := 1; o <= maxOrders; o++ {
				olCnt := rnd.Intn(11) + 5 // 5-15 order lines

				// Insert order
				_, err := g.db.Exec(orderQuery,
					o,                // o_id
					d,                // o_d_id
					w,                // o_w_id
					rnd.Intn(3000) + 1, // o_c_id
					time.Now(),       // o_entry_d
					rnd.Intn(10) + 1,  // o_carrier_id
					olCnt,            // o_ol_cnt
					1,                // o_all_local
				)
				if err != nil {
					return err
				}

				// Insert order lines
				for ol := 1; ol <= olCnt; ol++ {
					_, err = g.db.Exec(orderLineQuery,
						o,                // ol_o_id
						d,                // ol_d_id
						w,                // ol_w_id
						ol,               // ol_number
						rnd.Intn(100000) + 1, // ol_i_id
						w,                // ol_supply_w_id
						time.Now(),       // ol_delivery_d
						5,                // ol_quantity
						0.00,             // ol_amount
						"dist_info",      // ol_dist_info
					)
					if err != nil {
						return err
					}
				}

				// Insert new order for the last 900 orders
				if o > 2100 {
					_, err = g.db.Exec(newOrderQuery,
						o, // no_o_id
						d, // no_d_id
						w, // no_w_id
					)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// Helper functions

func (g *DataGenerator) generatePhoneNumber() string {
	return fmt.Sprintf("%03d-%03d-%03d-%04d",
		rand.Intn(1000),
		rand.Intn(1000),
		rand.Intn(1000),
		rand.Intn(10000))
}

func (g *DataGenerator) generateZip() string {
	return fmt.Sprintf("%04d11111", rand.Intn(10000))
}
