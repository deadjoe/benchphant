package benchmark

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"math/rand"
)

func TestNewDataGenerator(t *testing.T) {
	// 创建 mock 数据库连接
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	// 测试正常情况
	generator := NewDataGenerator(db, 10)
	assert.NotNil(t, generator)
	assert.Equal(t, 10, generator.warehouseCount)
	assert.NotNil(t, generator.db)

	// 测试边界情况
	generator = NewDataGenerator(db, 1)
	assert.NotNil(t, generator)
	assert.Equal(t, 1, generator.warehouseCount)

	// 验证 mock 的期望被满足
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestGenerateWarehouseData(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	generator := NewDataGenerator(db, 1)

	// 设置预期的 SQL 执行
	mock.ExpectExec("INSERT INTO warehouse").
		WithArgs(
			sqlmock.AnyArg(), // W_ID
			sqlmock.AnyArg(), // W_NAME
			sqlmock.AnyArg(), // W_STREET_1
			sqlmock.AnyArg(), // W_STREET_2
			sqlmock.AnyArg(), // W_CITY
			sqlmock.AnyArg(), // W_STATE
			sqlmock.AnyArg(), // W_ZIP
			sqlmock.AnyArg(), // W_TAX
			sqlmock.AnyArg(), // W_YTD
		).WillReturnResult(sqlmock.NewResult(1, 1))

	err = generator.GenerateWarehouseData()
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestGenerateDistrictData(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	generator := NewDataGenerator(db, 1)

	// 为每个仓库的10个区域设置预期
	for i := 0; i < 10; i++ {
		mock.ExpectExec("INSERT INTO district").
			WithArgs(
				sqlmock.AnyArg(), // D_ID
				sqlmock.AnyArg(), // D_W_ID
				sqlmock.AnyArg(), // D_NAME
				sqlmock.AnyArg(), // D_STREET_1
				sqlmock.AnyArg(), // D_STREET_2
				sqlmock.AnyArg(), // D_CITY
				sqlmock.AnyArg(), // D_STATE
				sqlmock.AnyArg(), // D_ZIP
				sqlmock.AnyArg(), // D_TAX
				sqlmock.AnyArg(), // D_YTD
				sqlmock.AnyArg(), // D_NEXT_O_ID
			).WillReturnResult(sqlmock.NewResult(1, 1))
	}

	err = generator.GenerateDistrictData()
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestGenerateCustomerData(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	generator := NewDataGenerator(db, 1)

	// 为每个区域的3000个客户设置预期
	for i := 0; i < 10; i++ { // 10个区域
		for j := 0; j < 3000; j++ { // 每个区域3000个客户
			// 设置客户数据插入的预期
			mock.ExpectExec("INSERT INTO customer").
				WithArgs(
					sqlmock.AnyArg(), // C_ID
					sqlmock.AnyArg(), // C_D_ID
					sqlmock.AnyArg(), // C_W_ID
					sqlmock.AnyArg(), // C_FIRST
					sqlmock.AnyArg(), // C_MIDDLE
					sqlmock.AnyArg(), // C_LAST
					sqlmock.AnyArg(), // C_STREET_1
					sqlmock.AnyArg(), // C_STREET_2
					sqlmock.AnyArg(), // C_CITY
					sqlmock.AnyArg(), // C_STATE
					sqlmock.AnyArg(), // C_ZIP
					sqlmock.AnyArg(), // C_PHONE
					sqlmock.AnyArg(), // C_SINCE
					sqlmock.AnyArg(), // C_CREDIT
					sqlmock.AnyArg(), // C_CREDIT_LIM
					sqlmock.AnyArg(), // C_DISCOUNT
					sqlmock.AnyArg(), // C_BALANCE
					sqlmock.AnyArg(), // C_YTD_PAYMENT
					sqlmock.AnyArg(), // C_PAYMENT_CNT
					sqlmock.AnyArg(), // C_DELIVERY_CNT
					sqlmock.AnyArg(), // C_DATA
				).WillReturnResult(sqlmock.NewResult(1, 1))

			// 设置历史记录插入的预期
			mock.ExpectExec("INSERT INTO history").
				WithArgs(
					sqlmock.AnyArg(), // H_C_ID
					sqlmock.AnyArg(), // H_C_D_ID
					sqlmock.AnyArg(), // H_C_W_ID
					sqlmock.AnyArg(), // H_D_ID
					sqlmock.AnyArg(), // H_W_ID
					sqlmock.AnyArg(), // H_DATE
					sqlmock.AnyArg(), // H_AMOUNT
					sqlmock.AnyArg(), // H_DATA
				).WillReturnResult(sqlmock.NewResult(1, 1))
		}
	}

	err = generator.GenerateCustomerData()
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestGenerateItemData(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	generator := NewDataGenerator(db, 1)

	// 设置物品数据插入的预期
	for i := 0; i < 100000; i++ {
		mock.ExpectExec("INSERT INTO item").
			WithArgs(
				sqlmock.AnyArg(), // I_ID
				sqlmock.AnyArg(), // I_IM_ID
				sqlmock.AnyArg(), // I_NAME
				sqlmock.AnyArg(), // I_PRICE
				sqlmock.AnyArg(), // I_DATA
			).WillReturnResult(sqlmock.NewResult(1, 1))
	}

	err = generator.GenerateItemData()
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestGenerateStockData(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	generator := NewDataGenerator(db, 1)

	// 设置库存数据插入的预期
	for i := 0; i < 100000; i++ {
		mock.ExpectExec("INSERT INTO stock").
			WithArgs(
				sqlmock.AnyArg(), // S_I_ID
				sqlmock.AnyArg(), // S_W_ID
				sqlmock.AnyArg(), // S_QUANTITY
				sqlmock.AnyArg(), // S_DIST_01
				sqlmock.AnyArg(), // S_DIST_02
				sqlmock.AnyArg(), // S_DIST_03
				sqlmock.AnyArg(), // S_DIST_04
				sqlmock.AnyArg(), // S_DIST_05
				sqlmock.AnyArg(), // S_DIST_06
				sqlmock.AnyArg(), // S_DIST_07
				sqlmock.AnyArg(), // S_DIST_08
				sqlmock.AnyArg(), // S_DIST_09
				sqlmock.AnyArg(), // S_DIST_10
				sqlmock.AnyArg(), // S_YTD
				sqlmock.AnyArg(), // S_ORDER_CNT
				sqlmock.AnyArg(), // S_REMOTE_CNT
				sqlmock.AnyArg(), // S_DATA
			).WillReturnResult(sqlmock.NewResult(1, 1))
	}

	err = generator.GenerateStockData()
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestGenerateOrderData(t *testing.T) {
	// 使用相同的随机数生成器
	rnd := rand.New(rand.NewSource(42))

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	generator := NewDataGenerator(db, 1)

	// 为了加快测试速度，我们只测试一个仓库的一个区域的前10个订单
	for w := 1; w <= 1; w++ {
		for d := 1; d <= 1; d++ {
			for o := 1; o <= 10; o++ {
				// 设置订单数据插入的预期
				olCnt := rnd.Intn(11) + 5 // 5-15 order lines
				mock.ExpectExec("INSERT INTO orders").
					WithArgs(
						o,                // o_id
						d,                // o_d_id
						w,                // o_w_id
						rnd.Intn(3000) + 1, // o_c_id
						sqlmock.AnyArg(), // o_entry_d
						rnd.Intn(10) + 1,  // o_carrier_id
						olCnt,            // o_ol_cnt
						1,                // o_all_local
					).WillReturnResult(sqlmock.NewResult(1, 1))

				// 设置订单行数据插入的预期
				for ol := 1; ol <= olCnt; ol++ {
					mock.ExpectExec("INSERT INTO order_line").
						WithArgs(
							o,                // ol_o_id
							d,                // ol_d_id
							w,                // ol_w_id
							ol,               // ol_number
							rnd.Intn(100000) + 1, // ol_i_id
							w,                // ol_supply_w_id
							sqlmock.AnyArg(), // ol_delivery_d
							5,                // ol_quantity
							0.00,             // ol_amount
							"dist_info",      // ol_dist_info
						).WillReturnResult(sqlmock.NewResult(1, 1))
				}

				// 为最后 900 个订单设置新订单数据插入的预期
				if o > 2100 {
					mock.ExpectExec("INSERT INTO new_order").
						WithArgs(
							o, // no_o_id
							d, // no_d_id
							w, // no_w_id
						).WillReturnResult(sqlmock.NewResult(1, 1))
				}
			}
		}
	}

	err = generator.GenerateOrderData()
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestDatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	generator := NewDataGenerator(db, 1)

	// 模拟数据库错误
	mock.ExpectExec("INSERT INTO warehouse").
		WillReturnError(sql.ErrConnDone)

	err = generator.GenerateWarehouseData()
	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
