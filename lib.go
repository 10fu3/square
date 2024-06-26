package square

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/10fu3/square/common"
	"github.com/10fu3/square/common/Operator"
	"github.com/10fu3/square/common/Orderby"
	"github.com/10fu3/square/common/where"
	"github.com/10fu3/square/lib"
	"strings"
)

type TableQuery struct {
	From     string
	Fields   ColumnsQuery
	Where    WhereQuery
	Limit    lib.Optional[uint]
	Offset   lib.Optional[uint]
	Orderby  []OrderbyQuery
	Relation *RelationTableQuery
	HasMany  bool
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
	Count   bool
}

type SelectColumnsQuery struct {
	Columns map[string]ColumnQuery
}

func BuildWhereOperator(w where.Op, table string) (string, []any) {
	switch w.Operator() {
	case Operator.True:
		return "true", []any{}
	case Operator.IsNull:
		return fmt.Sprintf("%s.%s IS NULL", table, w.ColumnName()), []any{}
	case Operator.In:
		inValues := make([]string, len(w.Args()))
		for i, _ := range w.Args() {
			inValues[i] = "?"
		}
		return fmt.Sprintf("%s.%s IN (%s)", table, w.ColumnName(), strings.Join(inValues, ", ")), w.Args()
	case Operator.NotIn:
		inValues := make([]string, len(w.Args()))
		for i, _ := range w.Args() {
			inValues[i] = "?"
		}
		return fmt.Sprintf("%s.%s NOT IN (%s)", table, w.ColumnName(), strings.Join(inValues, ", ")), w.Args()
	case Operator.Eq, Operator.Gt, Operator.Gte, Operator.Lt, Operator.Lte, Operator.Neq:
		return fmt.Sprintf("%s.%s %s ?", table, w.ColumnName(), w.Operator()), w.Args()
	}
	return "", []any{}
}

type WhereRelationQuery struct {
	//ParentTable   string
	ChildrenTable string
	Columns       []RelationColumn
	Where         *WhereQuery
}

type WhereQuery struct {
	joinFrom *struct {
		tableName string
		columns   []RelationColumn
	}
	Column   []where.Op
	Relation *WhereRelationQuery
	And      []WhereQuery
	Or       []WhereQuery
	Not      *WhereQuery
}

const wherequery = "(\nEXISTS \n" +
	"(SELECT 1 FROM\n" +
	" %s AS %s\n" +
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

func buildWhereQuery(w *WhereQuery, thisTable string) (string, []any) {
	var conditions = []string{"true"}
	var preparedStmt = []any{}

	if len(w.Column) > 0 {
		for _, column := range w.Column {
			columnConditions, columnPreparedStmt := BuildWhereOperator(column, thisTable)
			conditions = append(conditions, columnConditions)
			preparedStmt = append(preparedStmt, columnPreparedStmt...)
		}
	}

	if w.joinFrom != nil {
		relationships := []string{}
		for _, column := range w.joinFrom.columns {
			relationships = append(relationships, fmt.Sprintf(" ( \n%s.%s = %s.%s\n ) ", w.joinFrom.tableName, column.Parent, thisTable, column.This))
		}
		conditions = append(conditions, strings.Join(relationships, " AND "))
	}

	if w.Relation != nil {
		childrenWhereTable := common.MakeRandomStr()
		whereQuery, wherePreparedStmt := buildWhereQuery(w.Relation.Where, childrenWhereTable)
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
			andConditions, andPreparedStmt := buildWhereQuery(&and, thisTable)
			conditions = append(conditions, andConditions)
			preparedStmt = append(preparedStmt, andPreparedStmt...)
		}
	}

	if len(w.Or) > 0 {
		orConditions := []string{}
		for _, or := range w.Or {
			orConditions, orPreparedStmt := buildWhereQuery(&or, thisTable)
			conditions = append(conditions, orConditions)
			preparedStmt = append(preparedStmt, orPreparedStmt...)
		}
		conditions = append(conditions, fmt.Sprintf("(%s)", strings.Join(orConditions, " OR ")))
	}

	if w.Not != nil {
		notCondition, notPreparedStmt := buildWhereQuery(w.Not, thisTable)
		conditions = append(conditions, fmt.Sprintf("NOT (%s)", notCondition))
		preparedStmt = append(preparedStmt, notPreparedStmt...)
	}

	return strings.Join(conditions, " AND "), preparedStmt
}

type GenerateQuery string

var jsonArrayQuery = `SELECT
COALESCE(
    JSON_ARRAYAGG(__json__),
    CONVERT('[]', JSON)
) AS __data__
FROM
(%s) AS __data__`

var fieldJsonQuery = `JSON_OBJECT(%s) AS %s`

var jsonArrayElement = `SELECT
%s
FROM %s AS %s
WHERE %s
%s`

var jsonQuery = `SELECT
%s
FROM (%s) AS %s`

func buildOneResultQuery(q *TableQuery) (GenerateQuery, []any, error) {
	preparedStmt := []any{}
	jsonFields := []string{}
	dictInnerFields := map[string]struct{}{}
	innnerTableName := common.MakeRandomStr()
	for k, v := range q.Fields.Columns {
		if v.ColumnName == "" {
			v.ColumnName = k
		}
		jsonFields = append(jsonFields, fmt.Sprintf("'%s', %s.%s", k, innnerTableName, v.ColumnName))

		if v.Relation != nil {
			v.Relation.Where.joinFrom = &struct {
				tableName string
				columns   []RelationColumn
			}{
				tableName: innnerTableName,
				columns:   v.Relation.Relation.Columns,
			}
			refQuery, refQueryPstmt, err := BuildQuery(v.Relation)
			if err != nil {
				return "", nil, err
			}
			preparedStmt = append(preparedStmt, refQueryPstmt...)
			dictInnerFields[fmt.Sprintf("(%s) AS %s", refQuery, k)] = struct{}{}
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

	whereQuery, wherePstmt := buildWhereQuery(&q.Where, innnerTableName)

	innerSelect := fmt.Sprintf(
		jsonArrayElement,
		strings.Join(innerFields, ",\n"),
		q.From,
		innnerTableName,
		whereQuery,
		"", // not working group by
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
		1,
		q.Offset.GetOrDefault(0),
	)

	return GenerateQuery(elemQuery), preparedStmt, nil
}

func buildManyResultQuery(q *TableQuery) (GenerateQuery, []any, error) {
	//resultTableName := common.MakeRandomStr()
	preparedStmt := []any{}
	jsonFields := []string{}
	dictInnerFields := map[string]struct{}{}
	primitiveFields := []string{}

	groupByFields := []string{}
	groupByQuery := ""

	//outerTableName := common.MakeRandomStr()
	innnerTableName := common.MakeRandomStr()
	for k, v := range q.Fields.Columns {
		if v.ColumnName == "" {
			v.ColumnName = k
		}
		jsonFields = append(jsonFields, fmt.Sprintf("'%s', %s.%s", k, innnerTableName, v.ColumnName))

		if v.Relation != nil {
			v.Relation.Where.joinFrom = &struct {
				tableName string
				columns   []RelationColumn
			}{
				tableName: innnerTableName,
				columns:   v.Relation.Relation.Columns,
			}
			refQuery, pstmt, err := BuildQuery(v.Relation)
			if err != nil {
				return "", nil, err
			}
			dictInnerFields[fmt.Sprintf("(%s) AS %s", refQuery, k)] = struct{}{}
			preparedStmt = append(preparedStmt, pstmt...)
		} else {
			primitiveFields = append(primitiveFields, fmt.Sprintf(v.ColumnName))
			dictInnerFields[fmt.Sprintf("`%s`.`%s`", innnerTableName, v.ColumnName)] = struct{}{}
		}
	}
	for _, orderByColumn := range q.Orderby {
		dictInnerFields[fmt.Sprintf("`%s`.`%s`", innnerTableName, orderByColumn.Column)] = struct{}{}
	}

	if q.Fields.Count {
		jsonFields = append(jsonFields, fmt.Sprintf("'%s', %s.%s", "count", innnerTableName, "__count__"))
		dictInnerFields["COUNT(*) AS __count__"] = struct{}{}
		for _, k := range primitiveFields {
			groupByFields = append(groupByFields, k)
		}
		groupByQuery = fmt.Sprintf("GROUP BY %s", strings.Join(groupByFields, ", "))
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

	whereQuery, wherePstmt := buildWhereQuery(&q.Where, innnerTableName)

	innerSelect := fmt.Sprintf(
		jsonArrayElement,
		strings.Join(innerFields, ",\n"),
		q.From,
		innnerTableName,
		whereQuery,
		groupByQuery,
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

/**
 * BuildQuery: Build query from TableQuery
 * @param q *TableQuery
 * @return GeneratedQuery, PreparedStatement Variables, error
 */
func BuildQuery(q *TableQuery) (GenerateQuery, []any, error) {
	if q.HasMany {
		return buildManyResultQuery(q)
	}
	return buildOneResultQuery(q)
}

/**
 * FetchQuery: Fetch query from database
 * @param db *sql.DB
 * @param q *TableQuery
 * @param result *T
 * @return error
 */
func FetchQuery[T any](db *sql.DB, q *TableQuery, result *T) error {
	buildQuery, pstmt, e := BuildQuery(q)

	if e != nil {
		return e
	}

	stmt, err := db.Prepare(string(buildQuery))

	if err != nil {
		return err
	}

	rows, err := stmt.Query(pstmt...)

	if err != nil {
		return err
	}

	var resultRawExecute string
	for rows.Next() {
		rows.Scan(&resultRawExecute)
	}

	json.Unmarshal([]byte(resultRawExecute), &result)

	defer rows.Close()

	return nil
}

/**
 * FetchQueryTxn: Fetch query from database with transaction
 * @param db *sql.Tx
 * @param q *TableQuery
 * @param result *T
 * @return error
 */
func FetchQueryTxn[T any](db *sql.Tx, q *TableQuery, result *T) error {
	buildQuery, pstmt, e := BuildQuery(q)

	if e != nil {
		return e
	}

	stmt, err := db.Prepare(string(buildQuery))

	if err != nil {
		return err
	}

	rows, err := stmt.Query(pstmt...)

	if err != nil {
		return err
	}

	var resultRawExecute string
	for rows.Next() {
		rows.Scan(&resultRawExecute)
	}

	json.Unmarshal([]byte(resultRawExecute), &result)

	defer rows.Close()

	return nil
}
