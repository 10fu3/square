package common

import (
	"fmt"
	"github.com/10fu3/square/lib"
	"strings"
	"sync/atomic"
)

type Bool bool

func (b Bool) String() string {
	return fmt.Sprint(bool(b))
}

type BoolCompExp struct {
	Eq     lib.Optional[Bool]
	Neq    lib.Optional[Bool]
	IsNull lib.Optional[Bool]
}

func (exp *BoolCompExp) BuildQuery(preparedStatementOrder *atomic.Uint64, preparedStmtOrderAndValue map[uint64]any, field string) string {
	var conditions []string

	if exp.Eq.IsPresent() {
		preparedStmtOrderAndValue[preparedStatementOrder.Add(1)] = exp.Eq.Value
		conditions = append(conditions, fmt.Sprintf("(%s = ?)", field))
	}

	if exp.Neq.IsPresent() {
		preparedStmtOrderAndValue[preparedStatementOrder.Add(1)] = exp.Neq.Value
		conditions = append(conditions, fmt.Sprintf("(%s != ?)", field))
	}

	if exp.IsNull.IsPresent() {
		conditions = append(conditions, fmt.Sprintf("(%s IS NULL)", field))
	}

	return fmt.Sprintf("(%s)", strings.Join(conditions, " AND "))
}
