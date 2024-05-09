package common

import (
	"fmt"
	"github.com/10fu3/square/squrare/lib"
	"strings"
	"sync/atomic"
	"time"
)

type TimeCompExp struct {
	Eq     lib.Optional[time.Time]
	Gt     lib.Optional[time.Time]
	Gte    lib.Optional[time.Time]
	Lt     lib.Optional[time.Time]
	Lte    lib.Optional[time.Time]
	Neq    lib.Optional[time.Time]
	In     []time.Time
	NotIn  []time.Time
	IsNull lib.Optional[Bool]
}

func (exp *TimeCompExp) BuildQuery(preparedStatementOrder *atomic.Uint64, preparedStmtOrderAndValue map[uint64]any, field string) string {
	var conditions []string

	if exp.Eq.IsPresent() {
		preparedStmtOrderAndValue[preparedStatementOrder.Add(1)] = exp.Eq.Value.Format(time.DateTime)
		conditions = append(conditions, fmt.Sprintf("(%s = ?)", field))
	}

	if exp.Gt.IsPresent() {
		preparedStmtOrderAndValue[preparedStatementOrder.Add(1)] = exp.Gt.Value.Format(time.DateTime)
		conditions = append(conditions, fmt.Sprintf("(%s > ?)", field))
	}

	if exp.Gte.IsPresent() {
		preparedStmtOrderAndValue[preparedStatementOrder.Add(1)] = exp.Gte.Value.Format(time.DateTime)
		conditions = append(conditions, fmt.Sprintf("(%s >= ?)", field))
	}

	if exp.Lt.IsPresent() {
		preparedStmtOrderAndValue[preparedStatementOrder.Add(1)] = exp.Lt.Value.Format(time.DateTime)
		conditions = append(conditions, fmt.Sprintf("(%s < ?)", field))
	}

	if exp.Lte.IsPresent() {
		preparedStmtOrderAndValue[preparedStatementOrder.Add(1)] = exp.Lte.Value.Format(time.DateTime)
		conditions = append(conditions, fmt.Sprintf("(%s <= ?)", field))
	}

	if exp.Neq.IsPresent() {
		preparedStmtOrderAndValue[preparedStatementOrder.Add(1)] = exp.Neq.Value.Format(time.DateTime)
		conditions = append(conditions, fmt.Sprintf("(%s != ?)", field))
	}

	if len(exp.In) > 0 {
		expCondition := make([]string, len(exp.In))
		for i, v := range exp.In {
			expCondition[i] = "?"
			preparedStmtOrderAndValue[preparedStatementOrder.Add(1)] = v.Format(time.DateTime)
		}
		conditions = append(conditions, fmt.Sprintf("(%s IN (%s))", field, strings.Join(expCondition, ", ")))
	}

	if len(exp.NotIn) > 0 {
		expCondition := make([]string, len(exp.NotIn))
		for i, v := range exp.NotIn {
			expCondition[i] = "?"
			preparedStmtOrderAndValue[preparedStatementOrder.Add(1)] = v.Format(time.DateTime)
		}
		conditions = append(conditions, fmt.Sprintf("(%s NOT IN (%s))", field, strings.Join(expCondition, ", ")))
	}

	if exp.IsNull.IsPresent() {
		conditions = append(conditions, fmt.Sprintf("(%s IS NULL)", field))
	}

	return fmt.Sprintf("(%s)", strings.Join(conditions, " AND "))
}
