package common

import (
	"fmt"
	"github.com/10fu3/square/lib"
	"strings"
	"sync/atomic"
)

type Float float64

type FloatCompExp struct {
	Eq     lib.Optional[Float]
	Gt     lib.Optional[Float]
	Gte    lib.Optional[Float]
	Lt     lib.Optional[Float]
	Lte    lib.Optional[Float]
	Neq    lib.Optional[Float]
	In     []Float
	NotIn  []Float
	IsNull Bool
}

func (exp *FloatCompExp) BuildQuery(preparedStatementOrder *atomic.Uint64, preparedStmtOrderAndValue map[uint64]any, field string) string {
	var conditions []string

	if exp.Eq.IsPresent() {
		preparedStmtOrderAndValue[preparedStatementOrder.Add(1)] = exp.Eq.Value
		conditions = append(conditions, fmt.Sprintf("(%s = ?)", field))
	}

	if exp.Gt.IsPresent() {
		preparedStmtOrderAndValue[preparedStatementOrder.Add(1)] = exp.Gt.Value
		conditions = append(conditions, fmt.Sprintf("(%s > ?)", field))
	}

	if exp.Gte.IsPresent() {
		preparedStmtOrderAndValue[preparedStatementOrder.Add(1)] = exp.Gte.Value
		conditions = append(conditions, fmt.Sprintf("(%s >= ?)", field))
	}

	if exp.Lt.IsPresent() {
		preparedStmtOrderAndValue[preparedStatementOrder.Add(1)] = exp.Lt.Value
		conditions = append(conditions, fmt.Sprintf("(%s < ?)", field))
	}

	if exp.Lte.IsPresent() {
		preparedStmtOrderAndValue[preparedStatementOrder.Add(1)] = exp.Lte.Value
		conditions = append(conditions, fmt.Sprintf("(%s <= ?)", field))
	}

	if exp.Neq.IsPresent() {
		preparedStmtOrderAndValue[preparedStatementOrder.Add(1)] = exp.Neq.Value
		conditions = append(conditions, fmt.Sprintf("(%s != ?)", field))
	}

	if len(exp.In) > 0 {
		expCondition := make([]string, len(exp.In))
		for i, v := range exp.In {
			expCondition[i] = "?"
			preparedStmtOrderAndValue[preparedStatementOrder.Add(1)] = v
		}
		conditions = append(conditions, fmt.Sprintf("(%s IN (%s))", field, strings.Join(expCondition, ", ")))
	}

	if len(exp.NotIn) > 0 {
		expCondition := make([]string, len(exp.NotIn))
		for i, v := range exp.NotIn {
			expCondition[i] = "?"
			preparedStmtOrderAndValue[preparedStatementOrder.Add(1)] = v
		}
		conditions = append(conditions, fmt.Sprintf("(%s NOT IN (%s))", field, strings.Join(expCondition, ", ")))
	}

	if exp.IsNull {
		conditions = append(conditions, fmt.Sprintf("(%s IS NULL)", field))
	}

	return fmt.Sprintf("(%s)", strings.Join(conditions, " AND "))
}
