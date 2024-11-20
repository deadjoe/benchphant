package tpcc

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/deadjoe/benchphant/internal/benchmark"
)

// DatabaseConfig represents the database connection configuration
type DatabaseConfig struct {
	Type     string `json:"type"`     // mysql, postgresql
	Host     string `json:"host"`     // database host
	Port     int    `json:"port"`     // database port
	Username string `json:"username"` // database username
	Password string `json:"password"` // database password
	Database string `json:"database"` // database name
	SSLMode  string `json:"ssl_mode"` // disable, require, verify-ca, verify-full
}

// Config represents the TPC-C benchmark configuration
type Config struct {
	// Database configuration
	Database DatabaseConfig `json:"database"`

	// Scale configuration
	Warehouses int `json:"warehouses"` // Number of warehouses
	Terminals  int `json:"terminals"`  // Number of terminals (concurrent clients)

	// Duration configuration
	Duration       time.Duration `json:"duration"`        // Total test duration
	ReportInterval time.Duration `json:"report_interval"` // Interval between progress reports

	// Transaction mix configuration
	NewOrderPercentage    float64 `json:"new_order_percentage"`    // Percentage of new order transactions
	PaymentPercentage     float64 `json:"payment_percentage"`      // Percentage of payment transactions
	OrderStatusPercentage float64 `json:"order_status_percentage"` // Percentage of order status transactions
	DeliveryPercentage    float64 `json:"delivery_percentage"`     // Percentage of delivery transactions
	StockLevelPercentage  float64 `json:"stock_level_percentage"`  // Percentage of stock level transactions

	// New order configuration
	NewOrderItemsMin int `json:"new_order_items_min"` // Minimum items per new order
	NewOrderItemsMax int `json:"new_order_items_max"` // Maximum items per new order

	// Advanced configuration
	InitialLoad    bool `json:"initial_load"`    // Whether to load initial data
	DropExisting   bool `json:"drop_existing"`   // Whether to drop existing tables
	EnableForeign  bool `json:"enable_foreign"`  // Whether to enable foreign keys
	EnableIndexes  bool `json:"enable_indexes"`  // Whether to create indexes
	EnableTriggers bool `json:"enable_triggers"` // Whether to create triggers

	// Connection pool configuration
	MaxIdleConns    int           `json:"max_idle_conns"`    // Maximum number of idle connections
	MaxOpenConns    int           `json:"max_open_conns"`    // Maximum number of open connections
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"` // Maximum connection lifetime
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Warehouses <= 0 {
		return fmt.Errorf("warehouses must be greater than 0")
	}
	if c.Terminals <= 0 {
		return fmt.Errorf("terminals must be greater than 0")
	}
	if c.Duration <= 0 {
		return fmt.Errorf("duration must be greater than 0")
	}
	if c.ReportInterval <= 0 {
		return fmt.Errorf("report interval must be greater than 0")
	}

	// Validate transaction mix percentages
	total := c.NewOrderPercentage + c.PaymentPercentage + c.OrderStatusPercentage + c.DeliveryPercentage + c.StockLevelPercentage
	if total != 100 {
		return fmt.Errorf("transaction mix percentages must sum to 100, got %f", total)
	}

	// Validate new order items range
	if c.NewOrderItemsMin <= 0 || c.NewOrderItemsMax <= 0 {
		return fmt.Errorf("new order items min/max must be greater than 0")
	}
	if c.NewOrderItemsMin > c.NewOrderItemsMax {
		return fmt.Errorf("new order items min must be less than or equal to max")
	}

	// Validate connection pool settings
	if c.MaxIdleConns < 0 {
		return fmt.Errorf("max idle connections must be non-negative")
	}
	if c.MaxOpenConns < 0 {
		return fmt.Errorf("max open connections must be non-negative")
	}
	if c.ConnMaxLifetime < 0 {
		return fmt.Errorf("connection max lifetime must be non-negative")
	}

	return nil
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	switch c.Type {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			c.Username, c.Password, c.Host, c.Port, c.Database)
	case "postgresql":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.Username, c.Password, c.Database, c.SSLMode)
	default:
		return ""
	}
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Database: DatabaseConfig{
			Type:     "mysql",
			Host:     "localhost",
			Port:     3306,
			Username: "root",
			Password: "",
			Database: "tpcc",
			SSLMode:  "disable",
		},
		Warehouses:            10,
		Terminals:             100,
		Duration:              30 * time.Minute,
		ReportInterval:        1 * time.Minute,
		NewOrderPercentage:    45,
		PaymentPercentage:     43,
		OrderStatusPercentage: 4,
		DeliveryPercentage:    4,
		StockLevelPercentage:  4,
		NewOrderItemsMin:      5,
		NewOrderItemsMax:      15,
		InitialLoad:           true,
		DropExisting:          true,
		EnableForeign:         true,
		EnableIndexes:         true,
		EnableTriggers:        false,
		MaxIdleConns:          10,
		MaxOpenConns:          100,
		ConnMaxLifetime:       30 * time.Minute,
	}
}

// Stats represents TPCC test statistics
type Stats struct {
	TotalTransactions int64
	TPS              float64
	LatencyAvg       time.Duration
	LatencyP95       time.Duration
	LatencyP99       time.Duration
	Errors           int64
	StartTime        time.Time
	EndTime          time.Time
	Metrics          map[string]float64
}

// NewStats creates a new Stats instance
func NewStats() *Stats {
	return &Stats{
		StartTime: time.Now(),
		Metrics:   make(map[string]float64),
	}
}

// AddTransaction adds a transaction to the stats
func (s *Stats) AddTransaction(latency time.Duration, err error) {
	s.TotalTransactions++
	if err != nil {
		s.Errors++
		return
	}

	// Update latency metrics
	s.LatencyAvg = time.Duration((float64(s.LatencyAvg)*float64(s.TotalTransactions-1) + float64(latency)) / float64(s.TotalTransactions))
	
	// Update metrics map
	s.Metrics["total_transactions"] = float64(s.TotalTransactions)
	s.Metrics["errors"] = float64(s.Errors)
	s.Metrics["latency_avg_ms"] = float64(s.LatencyAvg.Milliseconds())
}

// Finalize calculates final statistics
func (s *Stats) Finalize() {
	s.EndTime = time.Now()
	duration := s.EndTime.Sub(s.StartTime)
	s.TPS = float64(s.TotalTransactions) / duration.Seconds()
	s.Metrics["tps"] = s.TPS
	s.Metrics["duration_seconds"] = duration.Seconds()
}

// Schema represents the TPC-C database schema
type Schema struct {
	Warehouse struct {
		ID      int     `json:"w_id"`
		Name    string  `json:"w_name"`
		Street1 string  `json:"w_street_1"`
		Street2 string  `json:"w_street_2"`
		City    string  `json:"w_city"`
		State   string  `json:"w_state"`
		Zip     string  `json:"w_zip"`
		Tax     float64 `json:"w_tax"`
		YTD     float64 `json:"w_ytd"`
	} `json:"warehouse"`

	District struct {
		ID      int     `json:"d_id"`
		WID     int     `json:"d_w_id"`
		Name    string  `json:"d_name"`
		Street1 string  `json:"d_street_1"`
		Street2 string  `json:"d_street_2"`
		City    string  `json:"d_city"`
		State   string  `json:"d_state"`
		Zip     string  `json:"d_zip"`
		Tax     float64 `json:"d_tax"`
		YTD     float64 `json:"d_ytd"`
		NextOID int     `json:"d_next_o_id"`
	} `json:"district"`

	Customer struct {
		ID           int     `json:"c_id"`
		DID          int     `json:"c_d_id"`
		WID          int     `json:"c_w_id"`
		First        string  `json:"c_first"`
		Middle       string  `json:"c_middle"`
		Last         string  `json:"c_last"`
		Street1      string  `json:"c_street_1"`
		Street2      string  `json:"c_street_2"`
		City         string  `json:"c_city"`
		State        string  `json:"c_state"`
		Zip          string  `json:"c_zip"`
		Phone        string  `json:"c_phone"`
		Since        string  `json:"c_since"`
		Credit       string  `json:"c_credit"`
		CreditLimit  float64 `json:"c_credit_lim"`
		Discount     float64 `json:"c_discount"`
		Balance      float64 `json:"c_balance"`
		YTDPayment   float64 `json:"c_ytd_payment"`
		PaymentCount int     `json:"c_payment_cnt"`
		DeliverCount int     `json:"c_delivery_cnt"`
		Data         string  `json:"c_data"`
	} `json:"customer"`

	Item struct {
		ID    int     `json:"i_id"`
		IMID  int     `json:"i_im_id"`
		Name  string  `json:"i_name"`
		Price float64 `json:"i_price"`
		Data  string  `json:"i_data"`
	} `json:"item"`

	Stock struct {
		ID        int    `json:"s_i_id"`
		WID       int    `json:"s_w_id"`
		Quantity  int    `json:"s_quantity"`
		Dist01    string `json:"s_dist_01"`
		Dist02    string `json:"s_dist_02"`
		Dist03    string `json:"s_dist_03"`
		Dist04    string `json:"s_dist_04"`
		Dist05    string `json:"s_dist_05"`
		Dist06    string `json:"s_dist_06"`
		Dist07    string `json:"s_dist_07"`
		Dist08    string `json:"s_dist_08"`
		Dist09    string `json:"s_dist_09"`
		Dist10    string `json:"s_dist_10"`
		YTD       int    `json:"s_ytd"`
		OrderCnt  int    `json:"s_order_cnt"`
		RemoteCnt int    `json:"s_remote_cnt"`
		Data      string `json:"s_data"`
	} `json:"stock"`

	Order struct {
		ID        int       `json:"o_id"`
		DID       int       `json:"o_d_id"`
		WID       int       `json:"o_w_id"`
		CID       int       `json:"o_c_id"`
		EntryD    time.Time `json:"o_entry_d"`
		CarrierID int       `json:"o_carrier_id"`
		OLCount   int       `json:"o_ol_cnt"`
		AllLocal  int       `json:"o_all_local"`
	} `json:"order"`

	NewOrder struct {
		OID int `json:"no_o_id"`
		DID int `json:"no_d_id"`
		WID int `json:"no_w_id"`
	} `json:"new_order"`

	OrderLine struct {
		OID       int       `json:"ol_o_id"`
		DID       int       `json:"ol_d_id"`
		WID       int       `json:"ol_w_id"`
		Number    int       `json:"ol_number"`
		IID       int       `json:"ol_i_id"`
		SupplyWID int       `json:"ol_supply_w_id"`
		DeliveryD time.Time `json:"ol_delivery_d"`
		Quantity  int       `json:"ol_quantity"`
		Amount    float64   `json:"ol_amount"`
		DistInfo  string    `json:"ol_dist_info"`
	} `json:"order_line"`

	History struct {
		CID    int       `json:"h_c_id"`
		CDID   int       `json:"h_c_d_id"`
		CWID   int       `json:"h_c_w_id"`
		DID    int       `json:"h_d_id"`
		WID    int       `json:"h_w_id"`
		Date   time.Time `json:"h_date"`
		Amount float64   `json:"h_amount"`
		Data   string    `json:"h_data"`
	} `json:"history"`
}

// Report represents a complete TPC-C test report
type Report struct {
	Config    Config    `json:"config"`
	Stats     Stats     `json:"stats"`
	Timestamp time.Time `json:"timestamp"`
	TestID    string    `json:"test_id"`
	TestName  string    `json:"test_name"`
}
