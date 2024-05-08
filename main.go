package main

import (
    "fmt"
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

func (w *WhereQueryColumn) BuildWhereQuery(table string, preparedStmt *PreparedStmtValues) string {
    switch w.Operator {
    case OperatorTrue:
        return "true"
    case OperatorIsNull:
        return fmt.Sprintf("%s.%s IS NULL", table, w.ColumnName)
    case OperatorIn:
        inValues := make([]string, len(w.Placeholder))
        for i, v := range w.Placeholder {
            p := common.MakeRandomStr()
            preparedStmt.values[p] = v
            inValues[i] = fmt.Sprintf(p)
        }
        return fmt.Sprintf("%s.%s IN (%s)", table, w.ColumnName, strings.Join(inValues, ", "))
    case OperatorNotIn:
        inValues := make([]string, len(w.Placeholder))
        for i, v := range w.Placeholder {
            p := common.MakeRandomStr()
            preparedStmt.values[p] = v
            inValues[i] = fmt.Sprintf(p)
        }
        return fmt.Sprintf("%s.%s NOT IN (%s)", table, w.ColumnName, strings.Join(inValues, ", "))
    case OperatorEq, OperatorGt, OperatorGte, OperatorLt, OperatorLte, OperatorNeq:
        p := common.MakeRandomStr()
        preparedStmt.values[p] = w.Placeholder[0]
        return fmt.Sprintf("%s.%s %s %s", table, w.ColumnName, w.Operator, p)
    }
    return ""
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

func BuildWhereQuery(w *WhereQuery, thisTable string, preparedStmt *PreparedStmtValues) string {
    var conditions = []string{"true"}

    if len(w.Column) > 0 {
        for _, column := range w.Column {
            conditions = append(conditions, column.BuildWhereQuery(thisTable, preparedStmt))
        }
    }

    if w.Relation != nil {
        childrenWhereTable := common.MakeRandomStr()
        whereQuery := BuildWhereQuery(w.Relation.Where, childrenWhereTable, preparedStmt)
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
    }

    if len(w.And) > 0 {
        for _, and := range w.And {
            conditions = append(conditions, BuildWhereQuery(&and, thisTable, preparedStmt))
        }
    }

    if len(w.Or) > 0 {
        orConditions := []string{}
        for _, or := range w.Or {
            orConditions = append(orConditions, BuildWhereQuery(&or, thisTable, preparedStmt))
        }
        conditions = append(conditions, fmt.Sprintf("(%s)", strings.Join(orConditions, " OR ")))
    }

    if w.Not != nil {
        notCondition := BuildWhereQuery(w.Not, thisTable, preparedStmt)
        conditions = append(conditions, fmt.Sprintf("NOT (%s)", notCondition))
    }

    return strings.Join(conditions, " AND ")
}

func FetchQuery(q TableQuery, defaultLimit uint, placeHolderValues []any) string {
    return ""
}

type TableName string

type GenerateQuery string
type PreparedStmtValues struct {
    values map[string]any
}

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

func buildOneResultQuery(q *TableQuery, pstmtv *PreparedStmtValues) (GenerateQuery, error) {
    //resultTableName := common.MakeRandomStr()
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
            refQuery, err := buildQuery(v.Relation, pstmtv)
            if err != nil {
                return "", err
            }
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

    innerSelect := fmt.Sprintf(
        jsonArrayElement,
        jsonSelect,
        q.From,
        innnerTableName,
        BuildWhereQuery(&q.Where, innnerTableName, pstmtv),
    )

    return GenerateQuery(innerSelect), nil
}

func buildManyResultQuery(q *TableQuery, pstmtv *PreparedStmtValues) (GenerateQuery, error) {
    //resultTableName := common.MakeRandomStr()
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
            refQuery, err := buildQuery(v.Relation, pstmtv)
            if err != nil {
                return "", err
            }
            dictInnerFields[fmt.Sprintf("(\n%s\n) AS %s", refQuery, v.ColumnName)] = struct{}{}
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

    innerSelect := fmt.Sprintf(
        jsonArrayElement,
        strings.Join(innerFields, ",\n"),
        q.From,
        innnerTableName,
        BuildWhereQuery(&q.Where, innnerTableName, pstmtv),
    )

    elemQuery := fmt.Sprintf(
        jsonQuery,
        jsonSelect,
        innerSelect,
        innnerTableName,
    )

    return GenerateQuery(fmt.Sprintf(jsonArrayQuery, elemQuery)), nil

}

func buildQuery(q *TableQuery, pstmtv *PreparedStmtValues) (GenerateQuery, error) {
    if q.IsFindOne {
        return buildOneResultQuery(q, pstmtv)
    }
    return buildManyResultQuery(q, pstmtv)
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
                                    },
                                },
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
    p := PreparedStmtValues{
        values: map[string]any{},
    }
    aaa, e := buildQuery(&q, &p)

    if e != nil {
        fmt.Println(e)
        return
    }

    fmt.Println(aaa)

}

func main() {
    a()
}
