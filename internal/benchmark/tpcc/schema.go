package tpcc

import (
	"context"
	"database/sql"
	"fmt"
)

// CreateSchema creates the TPC-C schema in the database
func CreateSchema(ctx context.Context, db *sql.DB) error {
	// Create tables in order of dependencies
	tables := []string{
		createWarehouseTable,
		createDistrictTable,
		createCustomerTable,
		createHistoryTable,
		createOrderTable,
		createNewOrderTable,
		createItemTable,
		createStockTable,
		createOrderLineTable,
	}

	// Create each table
	for _, table := range tables {
		if _, err := db.ExecContext(ctx, table); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	// Create indexes
	indexes := []string{
		createWarehouseIndex,
		createDistrictIndex,
		createCustomerIndex,
		createOrderIndex,
		createNewOrderIndex,
		createItemIndex,
		createStockIndex,
		createOrderLineIndex,
	}

	// Create each index
	for _, index := range indexes {
		if _, err := db.ExecContext(ctx, index); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// DropSchema drops all TPC-C tables and indexes
func DropSchema(ctx context.Context, db *sql.DB) error {
	tables := []string{
		"order_line",
		"stock",
		"item",
		"new_order",
		"orders",
		"history",
		"customer",
		"district",
		"warehouse",
	}

	for _, table := range tables {
		if _, err := db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	return nil
}

const (
	createWarehouseTable = `
CREATE TABLE warehouse (
	w_id        INTEGER            NOT NULL,
	w_name      VARCHAR(10)        NOT NULL,
	w_street_1  VARCHAR(20)        NOT NULL,
	w_street_2  VARCHAR(20)        NOT NULL,
	w_city      VARCHAR(20)        NOT NULL,
	w_state     CHAR(2)           NOT NULL,
	w_zip       CHAR(9)           NOT NULL,
	w_tax       DECIMAL(4,4)      NOT NULL,
	w_ytd       DECIMAL(12,2)     NOT NULL,
	PRIMARY KEY (w_id)
)`

	createDistrictTable = `
CREATE TABLE district (
	d_id         INTEGER         NOT NULL,
	d_w_id       INTEGER         NOT NULL,
	d_name       VARCHAR(10)     NOT NULL,
	d_street_1   VARCHAR(20)     NOT NULL,
	d_street_2   VARCHAR(20)     NOT NULL,
	d_city       VARCHAR(20)     NOT NULL,
	d_state      CHAR(2)        NOT NULL,
	d_zip        CHAR(9)        NOT NULL,
	d_tax        DECIMAL(4,4)   NOT NULL,
	d_ytd        DECIMAL(12,2)  NOT NULL,
	d_next_o_id  INTEGER        NOT NULL,
	PRIMARY KEY (d_w_id, d_id)
)`

	createCustomerTable = `
CREATE TABLE customer (
	c_id           INTEGER        NOT NULL,
	c_d_id         INTEGER        NOT NULL,
	c_w_id         INTEGER        NOT NULL,
	c_first        VARCHAR(16)    NOT NULL,
	c_middle       CHAR(2)        NOT NULL,
	c_last         VARCHAR(16)    NOT NULL,
	c_street_1     VARCHAR(20)    NOT NULL,
	c_street_2     VARCHAR(20)    NOT NULL,
	c_city         VARCHAR(20)    NOT NULL,
	c_state        CHAR(2)        NOT NULL,
	c_zip          CHAR(9)        NOT NULL,
	c_phone        CHAR(16)       NOT NULL,
	c_since        TIMESTAMP      NOT NULL,
	c_credit       CHAR(2)        NOT NULL,
	c_credit_lim   DECIMAL(12,2)  NOT NULL,
	c_discount     DECIMAL(4,4)   NOT NULL,
	c_balance      DECIMAL(12,2)  NOT NULL,
	c_ytd_payment  DECIMAL(12,2)  NOT NULL,
	c_payment_cnt  INTEGER        NOT NULL,
	c_delivery_cnt INTEGER        NOT NULL,
	c_data         VARCHAR(500)   NOT NULL,
	PRIMARY KEY (c_w_id, c_d_id, c_id)
)`

	createHistoryTable = `
CREATE TABLE history (
	h_c_id    INTEGER        NOT NULL,
	h_c_d_id  INTEGER        NOT NULL,
	h_c_w_id  INTEGER        NOT NULL,
	h_d_id    INTEGER        NOT NULL,
	h_w_id    INTEGER        NOT NULL,
	h_date    TIMESTAMP      NOT NULL,
	h_amount  DECIMAL(6,2)   NOT NULL,
	h_data    VARCHAR(24)    NOT NULL
)`

	createOrderTable = `
CREATE TABLE orders (
	o_id         INTEGER        NOT NULL,
	o_d_id       INTEGER        NOT NULL,
	o_w_id       INTEGER        NOT NULL,
	o_c_id       INTEGER        NOT NULL,
	o_entry_d    TIMESTAMP      NOT NULL,
	o_carrier_id INTEGER,
	o_ol_cnt     INTEGER        NOT NULL,
	o_all_local  INTEGER        NOT NULL,
	PRIMARY KEY (o_w_id, o_d_id, o_id)
)`

	createNewOrderTable = `
CREATE TABLE new_order (
	no_o_id  INTEGER    NOT NULL,
	no_d_id  INTEGER    NOT NULL,
	no_w_id  INTEGER    NOT NULL,
	PRIMARY KEY (no_w_id, no_d_id, no_o_id)
)`

	createItemTable = `
CREATE TABLE item (
	i_id     INTEGER        NOT NULL,
	i_im_id  INTEGER        NOT NULL,
	i_name   VARCHAR(24)    NOT NULL,
	i_price  DECIMAL(5,2)   NOT NULL,
	i_data   VARCHAR(50)    NOT NULL,
	PRIMARY KEY (i_id)
)`

	createStockTable = `
CREATE TABLE stock (
	s_i_id       INTEGER       NOT NULL,
	s_w_id       INTEGER       NOT NULL,
	s_quantity   INTEGER       NOT NULL,
	s_dist_01    CHAR(24)      NOT NULL,
	s_dist_02    CHAR(24)      NOT NULL,
	s_dist_03    CHAR(24)      NOT NULL,
	s_dist_04    CHAR(24)      NOT NULL,
	s_dist_05    CHAR(24)      NOT NULL,
	s_dist_06    CHAR(24)      NOT NULL,
	s_dist_07    CHAR(24)      NOT NULL,
	s_dist_08    CHAR(24)      NOT NULL,
	s_dist_09    CHAR(24)      NOT NULL,
	s_dist_10    CHAR(24)      NOT NULL,
	s_ytd        INTEGER       NOT NULL,
	s_order_cnt  INTEGER       NOT NULL,
	s_remote_cnt INTEGER       NOT NULL,
	s_data       VARCHAR(50)   NOT NULL,
	PRIMARY KEY (s_w_id, s_i_id)
)`

	createOrderLineTable = `
CREATE TABLE order_line (
	ol_o_id         INTEGER        NOT NULL,
	ol_d_id         INTEGER        NOT NULL,
	ol_w_id         INTEGER        NOT NULL,
	ol_number       INTEGER        NOT NULL,
	ol_i_id         INTEGER        NOT NULL,
	ol_supply_w_id  INTEGER        NOT NULL,
	ol_delivery_d   TIMESTAMP,
	ol_quantity     INTEGER        NOT NULL,
	ol_amount       DECIMAL(6,2)   NOT NULL,
	ol_dist_info    CHAR(24)       NOT NULL,
	PRIMARY KEY (ol_w_id, ol_d_id, ol_o_id, ol_number)
)`

	// Indexes
	createWarehouseIndex = `CREATE INDEX idx_warehouse ON warehouse (w_id)`
	createDistrictIndex  = `CREATE INDEX idx_district ON district (d_w_id, d_id)`
	createCustomerIndex  = `CREATE INDEX idx_customer ON customer (c_w_id, c_d_id, c_id)`
	createOrderIndex     = `CREATE INDEX idx_orders ON orders (o_w_id, o_d_id, o_id)`
	createNewOrderIndex  = `CREATE INDEX idx_new_order ON new_order (no_w_id, no_d_id, no_o_id)`
	createItemIndex      = `CREATE INDEX idx_item ON item (i_id)`
	createStockIndex     = `CREATE INDEX idx_stock ON stock (s_w_id, s_i_id)`
	createOrderLineIndex = `CREATE INDEX idx_order_line ON order_line (ol_w_id, ol_d_id, ol_o_id)`
)
