## example
### 1. create structured query
```go
structuredQuery := square.TableQuery{
    From: "video",
    Fields: square.ColumnsQuery{
        HasMany: true,
        Columns: map[string]square.ColumnQuery{
            "id":    {},
            "title": {},
            "video_actor": {
                Relation: &square.TableQuery{
                    HasMany: true,
                    Relation: &square.RelationTableQuery{
                        ParentName: "video",
                        ThisName:   "video_actor",
                        Columns: []square.RelationColumn{
                            {
                                Parent: "id",
                                This:   "video_id",
                            },
                        },
                    },
                    From: "video_actor",
                    Fields: square.ColumnsQuery{
                        Columns: map[string]square.ColumnQuery{
                            "video_id": {},
                            "actor": {
                                Relation: &square.TableQuery{
                                    From:      "actor",
                                    Fields: square.ColumnsQuery{
                                        Columns: map[string]square.ColumnQuery{
                                            "id":   {},
                                            "name": {},
                                        },
                                    },
                                    Relation: &square.RelationTableQuery{
                                        ParentName: "video_actor",
                                        ThisName:   "actor",
                                        Columns: []square.RelationColumn{
                                            {
                                                Parent: "actor_id",
                                                This:   "id",
                                            },
                                        },
                                    },
                                    Orderby: []square.OrderbyQuery{
                                        {
                                            Column: "id",
                                            Order:  Orderby.Desc,
                                        },
                                    },
                                },
                            },
                        },
                    },
                    Orderby: []square.OrderbyQuery{
                        {
                            Column: "video_id", 
                            Order:  Orderby.Desc,
                        },
                    },
                },
            },
        },
    },
    Where: square.WhereQuery{
        Column: []where.Op{
            where.Eq("id", 1),
        },
    },
    Offset: lib.NewOptional(uint(0)),
    Limit:  lib.NewOptional(uint(10)),
    Orderby: []square.OrderbyQuery{
        {
            Column: "id",
            Order:  Orderby.Desc,
        },
    },
}
```

### 2. generate prepared statement query and arguments
```go
createdQuery, arguments, err := square.BuildQuery(&structuredQuery)
```

### 3. (generated query)
```mysql
SELECT
    COALESCE(
            JSON_ARRAYAGG(__json__),
            CONVERT('[]', JSON)
    ) AS __data__
FROM
    (
        SELECT
            JSON_OBJECT(
                    'video_actor', JaGTdMwr.video_actor,
                    'id', JaGTdMwr.id, 'title', JaGTdMwr.title
            ) AS __json__
        FROM
            (
                SELECT
                    `JaGTdMwr`.`id`,
                    `JaGTdMwr`.`title`,
                    (
                        SELECT
                            COALESCE(
                                    JSON_ARRAYAGG(__json__),
                                    CONVERT('[]', JSON)
                            ) AS __data__
                        FROM
                            (
                                SELECT
                                    JSON_OBJECT(
                                            'video_id', MrcddPhl.video_id, 'actor',
                                            MrcddPhl.actor
                                    ) AS __json__
                                FROM
                                    (
                                        SELECT
                                            `MrcddPhl`.`video_id`,
                                            (
                                                SELECT
                                                    JSON_OBJECT(
                                                            'id', oWEhnsqu.id, 'name', oWEhnsqu.name
                                                    ) AS __json__
                                                FROM
                                                    (
                                                        SELECT
                                                            `oWEhnsqu`.`id`,
                                                            `oWEhnsqu`.`name`
                                                        FROM
                                                            actor AS oWEhnsqu
                                                        WHERE
                                                            true
                                                          AND (MrcddPhl.actor_id = oWEhnsqu.id)
                                                    ) AS oWEhnsqu
                                                ORDER BY
                                                    oWEhnsqu.id desc
                                                LIMIT
                                                    1 OFFSET 0
                                            ) AS actor
                                        FROM
                                            video_actor AS MrcddPhl
                                        WHERE
                                            true
                                          AND (JaGTdMwr.id = MrcddPhl.video_id)
                                    ) AS MrcddPhl
                                ORDER BY
                                    MrcddPhl.video_id desc
                                LIMIT
                                    9223372036854775807 
                                OFFSET 
                                    0
                            ) AS __data__
                    ) AS video_actor
                FROM
                    video AS JaGTdMwr
                WHERE
                    true
                  AND JaGTdMwr.id = ?
            ) AS JaGTdMwr
        ORDER BY
            JaGTdMwr.id desc
        LIMIT
            10 OFFSET 0
    ) AS __data__
```

### 3. execute query
```go
rows, err := db.QueryRow(createdQuery, arguments...)
```
### 4. scan rows
```go
var data string
err = rows.Scan(&data)
```

### 5. unmarshal json
```go
var result []struct {
	Id         int    `json:"id"`
	VideoTitle string `json:"title"`
	VideoActor []struct {
		VideoId int `json:"video_id"`
		Actor   struct {
			Id   int    `json:"id"`
			Name string `json:"name"`
		} `json:"actor"`
	} `json:"video_actor"`
}

err = json.Unmarshal([]byte(data), &result)
```

