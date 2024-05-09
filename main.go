package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"square/common"
	"square/common/Orderby"
	"square/lib"
	"strings"
)

type TableQuery struct {
	From      string
	Fields    ColumnsQuery
	Where     WhereQuery
	Limit     lib.Optional[uint]
	Offset    lib.Optional[uint]
	Orderby   []OrderbyQuery
	Relation  *RelationTableQuery
	IsFindOne bool
}

type RelationTableQuery struct {
	ParentName string
	ThisName   string
	Columns    []RelationColumn
}

type RelationColumn struct {
	Parent string
	This   string
}

type OrderbyQuery struct {
	Column string
	Order  Orderby.OrderBy
}

type ColumnQuery struct {
	Relation   *TableQuery
	ColumnName string
}

type ColumnsQuery struct {
	Columns map[string]ColumnQuery
}

type SelectColumnsQuery struct {
	Columns map[string]ColumnQuery
}

type WhereQueryOperator uint

const (
	OperatorTrue   WhereQueryOperator = iota
	OperatorEq     WhereQueryOperator = iota
	OperatorGt     WhereQueryOperator = iota
	OperatorGte    WhereQueryOperator = iota
	OperatorLt     WhereQueryOperator = iota
	OperatorLte    WhereQueryOperator = iota
	OperatorNeq    WhereQueryOperator = iota
	OperatorIn     WhereQueryOperator = iota
	OperatorNotIn  WhereQueryOperator = iota
	OperatorIsNull WhereQueryOperator = iota
)

func (o WhereQueryOperator) String() string {
	switch o {
	case OperatorEq:
		return "="
	case OperatorGt:
		return ">"
	case OperatorGte:
		return ">="
	case OperatorLt:
		return "<"
	case OperatorLte:
		return "<="
	case OperatorNeq:
		return "!="
	case OperatorIn:
		return "IN"
	case OperatorNotIn:
		return "NOT IN"
	case OperatorIsNull:
		return "IS NULL"
	default:
		return "1 = 1"
	}
}

type WhereQueryColumn struct {
	ColumnName    string
	Placeholder   []any
	ConstantValue string
	Operator      WhereQueryOperator
}

func (w *WhereQueryColumn) BuildWhereQuery(table string) (string, []any) {
	switch w.Operator {
	case OperatorTrue:
		return "true", []any{}
	case OperatorIsNull:
		return fmt.Sprintf("%s.%s IS NULL", table, w.ColumnName), []any{}
	case OperatorIn:
		inValues := make([]string, len(w.Placeholder))
		for i, _ := range w.Placeholder {
			inValues[i] = "?"
		}
		return fmt.Sprintf("%s.%s IN (%s)", table, w.ColumnName, strings.Join(inValues, ", ")), w.Placeholder
	case OperatorNotIn:
		inValues := make([]string, len(w.Placeholder))
		for i, _ := range w.Placeholder {
			inValues[i] = "?"
		}
		return fmt.Sprintf("%s.%s NOT IN (%s)", table, w.ColumnName, strings.Join(inValues, ", ")), w.Placeholder
	case OperatorEq, OperatorGt, OperatorGte, OperatorLt, OperatorLte, OperatorNeq:
		return fmt.Sprintf("%s.%s %s ?", table, w.ColumnName, w.Operator), w.Placeholder
	}
	return "", []any{}
}

type WhereRelationQuery struct {
	ParentTable   string
	ChildrenTable string
	Columns       []RelationColumn
	Where         *WhereQuery
}

type WhereQuery struct {
	Column   []WhereQueryColumn
	Relation *WhereRelationQuery
	And      []WhereQuery
	Or       []WhereQuery
	Not      *WhereQuery
}

const wherequery = "(\nEXISTS \n" +
	"(SELECT 1 FROM %s AS %s\n" +
	" WHERE \n" +
	" %s\n" +
	" AND %s\n)\n)"

func GenerateWhereQuery(
	childTableName,
	childTableRandomName,
	relationships,
	andmore string) string {
	return fmt.Sprintf(wherequery,
		childTableName,
		childTableRandomName,
		relationships,
		andmore)
}

func BuildWhereQuery(w *WhereQuery, thisTable string) (string, []any) {
	var conditions = []string{"true"}
	var preparedStmt = []any{}

	if len(w.Column) > 0 {
		for _, column := range w.Column {
			columnConditions, columnPreparedStmt := column.BuildWhereQuery(thisTable)
			conditions = append(conditions, columnConditions)
			preparedStmt = append(preparedStmt, columnPreparedStmt...)
		}
	}

	if w.Relation != nil {
		childrenWhereTable := common.MakeRandomStr()
		whereQuery, wherePreparedStmt := BuildWhereQuery(w.Relation.Where, childrenWhereTable)
		relationships := []string{}
		for _, column := range w.Relation.Columns {
			relationships = append(relationships, fmt.Sprintf(" ( \n%s.%s = %s.%s\n ) ", thisTable, column.Parent, childrenWhereTable, column.This))
		}
		conditions = append(conditions, GenerateWhereQuery(
			w.Relation.ChildrenTable,
			childrenWhereTable,
			strings.Join(relationships, " AND "),
			whereQuery,
		))
		preparedStmt = append(preparedStmt, wherePreparedStmt...)
	}

	if len(w.And) > 0 {
		for _, and := range w.And {
			andConditions, andPreparedStmt := BuildWhereQuery(&and, thisTable)
			conditions = append(conditions, andConditions)
			preparedStmt = append(preparedStmt, andPreparedStmt...)
		}
	}

	if len(w.Or) > 0 {
		orConditions := []string{}
		for _, or := range w.Or {
			orConditions, orPreparedStmt := BuildWhereQuery(&or, thisTable)
			conditions = append(conditions, orConditions)
			preparedStmt = append(preparedStmt, orPreparedStmt...)
		}
		conditions = append(conditions, fmt.Sprintf("(%s)", strings.Join(orConditions, " OR ")))
	}

	if w.Not != nil {
		notCondition, notPreparedStmt := BuildWhereQuery(w.Not, thisTable)
		conditions = append(conditions, fmt.Sprintf("NOT (%s)", notCondition))
		preparedStmt = append(preparedStmt, notPreparedStmt...)
	}

	return strings.Join(conditions, " AND "), preparedStmt
}

type GenerateQuery string

var jsonArrayQuery = `
SELECT
    COALESCE(
        JSON_ARRAYAGG(__json__),
        CONVERT('[]', JSON)
    ) AS __data__
FROM
(
    %s
) AS __data__
`

var fieldJsonQuery = `
JSON_OBJECT(
    %s
) AS %s
`

var jsonArrayElement = `
SELECT
    %s
FROM %s AS %s
WHERE %s
`

var jsonQuery = `
SELECT
    %s
FROM (
    %s
) AS %s
`

func buildOneResultQuery(q *TableQuery) (GenerateQuery, []any, error) {
	//resultTableName := common.MakeRandomStr()
	preparedStmt := []any{}
	jsonFields := []string{}
	dictInnerFields := map[string]struct{}{}
	//outerTableName := common.MakeRandomStr()
	innnerTableName := common.MakeRandomStr()
	for k, v := range q.Fields.Columns {
		if v.ColumnName == "" {
			v.ColumnName = k
		}
		jsonFields = append(jsonFields, fmt.Sprintf("'%s', %s.%s", v.ColumnName, innnerTableName, v.ColumnName))

		if v.Relation != nil {
			refQuery, refQueryPstmt, err := BuildQuery(v.Relation)
			if err != nil {
				return "", nil, err
			}
			preparedStmt = append(preparedStmt, refQueryPstmt...)
			dictInnerFields[fmt.Sprintf("(\n%s\n) AS %s", refQuery, v.ColumnName)] = struct{}{}
		} else {
			dictInnerFields[fmt.Sprintf("`%s`.`%s`", innnerTableName, v.ColumnName)] = struct{}{}
		}
	}
	innerFields := []string{}
	for k := range dictInnerFields {
		innerFields = append(innerFields, k)
	}

	jsonSelect := fmt.Sprintf(
		fieldJsonQuery,
		strings.Join(jsonFields, ",\n"),
		"__json__",
	)

	whereQuery, whrePstmtv := BuildWhereQuery(&q.Where, innnerTableName)

	innerSelect := fmt.Sprintf(
		jsonArrayElement,
		jsonSelect,
		q.From,
		innnerTableName,
		whereQuery,
	)

	preparedStmt = append(preparedStmt, whrePstmtv...)

	return GenerateQuery(innerSelect), preparedStmt, nil
}

func buildManyResultQuery(q *TableQuery) (GenerateQuery, []any, error) {
	//resultTableName := common.MakeRandomStr()
	preparedStmt := []any{}
	jsonFields := []string{}
	dictInnerFields := map[string]struct{}{}
	//outerTableName := common.MakeRandomStr()
	innnerTableName := common.MakeRandomStr()
	for k, v := range q.Fields.Columns {
		if v.ColumnName == "" {
			v.ColumnName = k
		}
		jsonFields = append(jsonFields, fmt.Sprintf("'%s', %s.%s", v.ColumnName, innnerTableName, v.ColumnName))

		if v.Relation != nil {
			refQuery, pstmt, err := BuildQuery(v.Relation)
			if err != nil {
				return "", nil, err
			}
			dictInnerFields[fmt.Sprintf("(\n%s\n) AS %s", refQuery, v.ColumnName)] = struct{}{}
			preparedStmt = append(preparedStmt, pstmt...)
		} else {
			dictInnerFields[fmt.Sprintf("`%s`.`%s`", innnerTableName, v.ColumnName)] = struct{}{}
		}
	}
	for _, orderByColumn := range q.Orderby {
		dictInnerFields[fmt.Sprintf("`%s`.`%s`", innnerTableName, orderByColumn.Column)] = struct{}{}
	}
	innerFields := []string{}
	for k := range dictInnerFields {
		innerFields = append(innerFields, k)
	}

	jsonSelect := fmt.Sprintf(
		fieldJsonQuery,
		strings.Join(jsonFields, ",\n"),
		"__json__",
	)

	whereQuery, wherePstmt := BuildWhereQuery(&q.Where, innnerTableName)

	innerSelect := fmt.Sprintf(
		jsonArrayElement,
		strings.Join(innerFields, ",\n"),
		q.From,
		innnerTableName,
		whereQuery,
	)

	preparedStmt = append(preparedStmt, wherePstmt...)

	elemQuery := fmt.Sprintf(
		jsonQuery,
		jsonSelect,
		innerSelect,
		innnerTableName,
	)

	orderByQuery := []string{}
	for _, orderByColumn := range q.Orderby {
		orderByQuery = append(orderByQuery, fmt.Sprintf("%s.%s %s", innnerTableName, orderByColumn.Column, orderByColumn.Order))
	}

	if len(orderByQuery) > 0 {
		elemQuery = fmt.Sprintf(
			"%s\nORDER BY %s",
			elemQuery,
			strings.Join(orderByQuery, ", "),
		)
	}

	elemQuery = fmt.Sprintf(
		"%s\nLIMIT %d OFFSET %d",
		elemQuery,
		q.Limit.GetOrDefault(9223372036854775807),
		q.Offset.GetOrDefault(0),
	)

	return GenerateQuery(fmt.Sprintf(jsonArrayQuery, elemQuery)), preparedStmt, nil

}

func BuildQuery(q *TableQuery) (GenerateQuery, []any, error) {
	if q.IsFindOne {
		return buildOneResultQuery(q)
	}
	return buildManyResultQuery(q)
}

func a() {
	q := TableQuery{
		From: "video",
		Fields: ColumnsQuery{
			Columns: map[string]ColumnQuery{
				"id":    {},
				"title": {},
				"video_actor": {
					Relation: &TableQuery{
						Relation: &RelationTableQuery{
							ParentName: "video",
							ThisName:   "video_actor",
							Columns: []RelationColumn{
								{
									Parent: "id",
									This:   "video_id",
								},
							},
						},
						From: "video_actor",
						Fields: ColumnsQuery{
							Columns: map[string]ColumnQuery{
								"video_id": {},
								"actor": {
									Relation: &TableQuery{
										From: "actor",
										Fields: ColumnsQuery{
											Columns: map[string]ColumnQuery{
												"id":   {},
												"name": {},
											},
										},
										Relation: &RelationTableQuery{
											ParentName: "video_actor",
											ThisName:   "actor",
											Columns: []RelationColumn{
												{
													Parent: "actor_id",
													This:   "id",
												},
											},
										},
										Orderby: []OrderbyQuery{
											{
												Column: "id",
												Order:  Orderby.Desc,
											},
										},
									},
								},
							},
						},
						Orderby: []OrderbyQuery{
							{
								Column: "video_id",
								Order:  Orderby.Desc,
							},
						},
					},
				},
			},
		},
		Where: WhereQuery{
			Column: []WhereQueryColumn{
				{
					ColumnName:  "id",
					Operator:    OperatorEq,
					Placeholder: []any{1},
				},
			},
			Relation: &WhereRelationQuery{
				ParentTable:   "video",
				ChildrenTable: "video_actor",
				Columns: []RelationColumn{
					{
						Parent: "id",
						This:   "video_id",
					},
				},
				Where: &WhereQuery{
					Column: []WhereQueryColumn{
						{
							ColumnName:  "actor_id",
							Operator:    OperatorEq,
							Placeholder: []any{1},
						},
					},
				},
			},
		},
		Offset: lib.Optional[uint]{
			Value: 0,
		},
		Limit: lib.Optional[uint]{
			Value: 10,
		},
		Orderby: []OrderbyQuery{
			{
				Column: "id",
				Order:  Orderby.Desc,
			},
		},
	}

	aaa, pstmt, e := BuildQuery(&q)

	if e != nil {
		panic(e)
	}

	// connection db
	conn, err := sql.Open("mysql", "root:root@tcp(localhost:3306)/video_service?charset=utf8&parseTime=true")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	stmt, err := conn.Prepare(string(aaa))

	if err != nil {
		panic(err)
	}

	rows, err := stmt.Query(pstmt...)

	if err != nil {
		panic(err)
	}
	line := ""
	for rows.Next() {
		rows.Scan(&line)
	}

	var x interface{}
	json.Unmarshal([]byte(line), &x)

	b, err := json.MarshalIndent(x, "", "  ")

	fmt.Println(string(b))

	var y []struct {
		Id         int    `json:"id"`
		Title      string `json:"title"`
		VideoActor []struct {
			VideoId int `json:"video_id"`
			Actor   []struct {
				Id   int    `json:"id"`
				Name string `json:"name"`
			} `json:"actor"`
		} `json:"video_actor"`
	}

	json.Unmarshal([]byte(line), &y)

	fmt.Println(y)

	defer rows.Close()
}

func main() {
	a()
}
