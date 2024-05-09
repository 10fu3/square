package main

import (
	"database/sql"
	"fmt"
	"github.com/10fu3/square/common/Operator"
	"github.com/10fu3/square/common/Orderby"
	"github.com/10fu3/square/lib"
	_ "github.com/go-sql-driver/mysql"
)

func a() {
	structuredQuery := TableQuery{
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
												"id": {},
												"actor_name": {
													ColumnName: "name",
												},
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
					Operator:    Operator.Eq,
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
							Operator:    Operator.Eq,
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

	var result []struct {
		Id         int    `json:"id"`
		VideoTitle string `json:"video_title"`
		VideoActor []struct {
			VideoId int `json:"video_id"`
			Actor   []struct {
				Id        int    `json:"id"`
				ActorName string `json:"actor_name"`
			} `json:"actor"`
		} `json:"video_actor"`
	}

	// connection db
	conn, err := sql.Open("mysql", "root:root@tcp(localhost:3306)/video_service?charset=utf8&parseTime=true")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	err = FetchQuery(conn, &structuredQuery, &result)

	if err != nil {
		panic(err)
	}

	fmt.Println(result)
}

func main() {
	a()
}
