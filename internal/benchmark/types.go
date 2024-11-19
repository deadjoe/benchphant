package benchmark

// BenchmarkType represents the type of benchmark
type BenchmarkType string

const (
	// BenchmarkTypeTPCC represents TPC-C benchmark
	BenchmarkTypeTPCC BenchmarkType = "tpcc"
	// BenchmarkTypeTPCH represents TPC-H benchmark
	BenchmarkTypeTPCH BenchmarkType = "tpch"
	// BenchmarkTypeYCSB represents YCSB benchmark
	BenchmarkTypeYCSB BenchmarkType = "ycsb"
)
