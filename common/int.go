package common

import (
    "fmt"
    "square/lib"

    "strings"
    "sync/atomic"
)

type Uint uint

func (u Uint) String() string {
    return fmt.Sprint(uint(u))
}

type Int int

type IntCompExp struct {
    Eq     lib.Optional[Int]
    Gt     lib.Optional[Int]
    Gte    lib.Optional[Int]
    Lt     lib.Optional[Int]
    Lte    lib.Optional[Int]
    Neq    lib.Optional[Int]
    In     []Int
    NotIn  []Int
    IsNull bool
}

func (exp *IntCompExp) BuildQuery(preparedStatementOrder *atomic.Uint64, preparedStmtOrderAndValue map[uint64]any, field string) string {
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
