package common

import (
	"fmt"
	"github.com/10fu3/square/lib"
	"strings"
	"sync/atomic"
)

type Str string

type StrCompExp struct {
	Eq      lib.Optional[string]
	Gt      lib.Optional[string]
	Gte     lib.Optional[string]
	Lt      lib.Optional[string]
	Lte     lib.Optional[string]
	Neq     lib.Optional[string]
	Like    lib.Optional[string]
	NotLike lib.Optional[string]
	Ilike   lib.Optional[string]
	In      []string
	NotIn   []string
	IsNull  lib.Optional[bool]
}

func (exp *StrCompExp) BuildQuery(preparedStatementOrder *atomic.Uint64, preparedStmtOrderAndValue map[uint64]any, field string) string {
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

	if exp.IsNull.IsPresent() {
		conditions = append(conditions, fmt.Sprintf("(%s IS NULL)", field))
	}

	return fmt.Sprintf("(%s)", strings.Join(conditions, " AND "))
}
