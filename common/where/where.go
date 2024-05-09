package where

import "github.com/10fu3/square/common/Operator"

type Op interface {
	ColumnName() string
	Operator() Operator.Type
	Args() []any
}

type whereIsNullOperator struct {
	columnName string
}

func (w *whereIsNullOperator) ColumnName() string {
	return w.columnName
}

func (w *whereIsNullOperator) Operator() Operator.Type {
	return Operator.IsNull
}

func (w *whereIsNullOperator) Args() []any {
	return []any{}
}

func IsNull(columnName string) Op {
	return &whereIsNullOperator{
		columnName: columnName,
	}
}

type whereInOperator struct {
	columnName string
	args       []any
}

func (w *whereInOperator) ColumnName() string {
	return w.columnName
}

func (w *whereInOperator) Operator() Operator.Type {
	return Operator.In
}

func (w *whereInOperator) Args() []any {
	return w.args
}

func In(columnName string, args ...any) Op {
	return &whereInOperator{
		columnName: columnName,
		args:       args,
	}
}

type whereTrueOperator struct{}

func (w *whereTrueOperator) ColumnName() string {
	return ""
}

func (w *whereTrueOperator) Operator() Operator.Type {
	return Operator.True
}

func (w *whereTrueOperator) Args() []any {
	return []any{}
}

func True() Op {
	return &whereTrueOperator{}
}

type whereNotInOperator struct {
	columnName string
	args       []any
}

func (w *whereNotInOperator) ColumnName() string {
	return w.columnName
}

func (w *whereNotInOperator) Operator() Operator.Type {
	return Operator.NotIn
}

func (w *whereNotInOperator) Args() []any {
	return w.args
}

func NotIn(columnName string, args ...any) Op {
	return &whereNotInOperator{
		columnName: columnName,
		args:       args,
	}
}

type whereEqOperator struct {
	columnName string
	args       []any
}

func (w *whereEqOperator) ColumnName() string {
	return w.columnName
}

func (w *whereEqOperator) Operator() Operator.Type {
	return Operator.Eq
}

func (w *whereEqOperator) Args() []any {
	return w.args
}

func Eq(columnName string, args any) Op {
	return &whereEqOperator{
		columnName: columnName,
		args:       []any{args},
	}
}

type whereGtOperator struct {
	columnName string
	args       []any
}

func (w *whereGtOperator) ColumnName() string {
	return w.columnName
}

func (w *whereGtOperator) Operator() Operator.Type {
	return Operator.Gt
}

func (w *whereGtOperator) Args() []any {
	return w.args
}

func Gt(columnName string, args any) Op {
	return &whereGtOperator{
		columnName: columnName,
		args:       []any{args},
	}
}

type whereGteOperator struct {
	columnName string
	args       []any
}

func (w *whereGteOperator) ColumnName() string {
	return w.columnName
}

func (w *whereGteOperator) Operator() Operator.Type {
	return Operator.Gte
}

func (w *whereGteOperator) Args() []any {
	return w.args
}

func Gte(columnName string, args any) Op {
	return &whereGteOperator{
		columnName: columnName,
		args:       []any{args},
	}
}

type whereLtOperator struct {
	columnName string
	args       []any
}

func (w *whereLtOperator) ColumnName() string {
	return w.columnName
}

func (w *whereLtOperator) Operator() Operator.Type {
	return Operator.Lt
}

func (w *whereLtOperator) Args() []any {
	return w.args
}

func Lt(columnName string, args any) Op {
	return &whereLtOperator{
		columnName: columnName,
		args:       []any{args},
	}
}

type whereLteOperator struct {
	columnName string
	args       []any
}

func (w *whereLteOperator) ColumnName() string {
	return w.columnName
}

func (w *whereLteOperator) Operator() Operator.Type {
	return Operator.Lte
}

func (w *whereLteOperator) Args() []any {
	return w.args
}

func Lte(columnName string, args any) Op {
	return &whereLteOperator{
		columnName: columnName,
		args:       []any{args},
	}
}

type whereNeqOperator struct {
	columnName string
	args       []any
}

func (w *whereNeqOperator) ColumnName() string {
	return w.columnName
}

func (w *whereNeqOperator) Operator() Operator.Type {
	return Operator.Neq
}

func (w *whereNeqOperator) Args() []any {
	return w.args
}

func Neq(columnName string, args any) Op {
	return &whereNeqOperator{
		columnName: columnName,
		args:       []any{args},
	}
}
