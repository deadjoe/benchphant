package benchmark

import (
	"fmt"
	"math/rand"
	"time"
)

// QueryDistributionType represents the type of query distribution
type QueryDistributionType string

const (
	// QueryDistributionRandom executes queries in random order
	QueryDistributionRandom QueryDistributionType = "random"
	// QueryDistributionWeighted executes queries based on weights
	QueryDistributionWeighted QueryDistributionType = "weighted"
)

// QueryType represents the type of a query
type QueryType string

// Query represents a SQL query
type Query struct {
	SQL      string    `json:"sql"`
	Type     QueryType `json:"type"`
	Name     string    `json:"name"`
	Prepared bool      `json:"prepared"`
}

// QueryExecutor executes queries according to a distribution strategy
type QueryExecutor struct {
	queries      []string
	distribution QueryDistributionType
	weights      []float64
	current      int
	rnd          *rand.Rand
}

// NewQueryExecutor creates a new query executor
func NewQueryExecutor(queries []string, distribution QueryDistributionType, weights []float64) (*QueryExecutor, error) {
	if len(queries) == 0 {
		return nil, fmt.Errorf("no queries provided")
	}

	if distribution == QueryDistributionWeighted && (weights == nil || len(weights) != len(queries)) {
		return nil, fmt.Errorf("weights must be provided for weighted distribution")
	}

	return &QueryExecutor{
		queries:      queries,
		distribution: distribution,
		weights:      weights,
		rnd:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

// NextQuery returns the next query to execute
func (e *QueryExecutor) NextQuery() string {
	switch e.distribution {
	case QueryDistributionRandom:
		return e.queries[e.rnd.Intn(len(e.queries))]

	case QueryDistributionWeighted:
		// Select a query based on weights
		r := e.rnd.Float64()
		var sum float64
		for i, w := range e.weights {
			sum += w
			if r <= sum {
				return e.queries[i]
			}
		}
		// Fallback to last query if we somehow got here
		return e.queries[len(e.queries)-1]

	default:
		// Default to random
		return e.queries[e.rnd.Intn(len(e.queries))]
	}
}
