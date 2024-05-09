package conn_test

import (
	"database/sql"
	"fmt"
	"github.com/10fu3/square"
	"github.com/10fu3/square/common/Operator"
	"github.com/10fu3/square/common/Orderby"
	"github.com/10fu3/square/lib"
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

func TestCoverArray(t *testing.T) {
	structuredQuery := square.TableQuery{
		From: "video",
		Fields: square.ColumnsQuery{
			Columns: map[string]square.ColumnQuery{
				"id":    {},
				"title": {},
				"video_actor": {
					Relation: &square.TableQuery{
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
										From: "actor",
										Fields: square.ColumnsQuery{
											Columns: map[string]square.ColumnQuery{
												"id": {},
												"actor_name": {
													ColumnName: "name",
												},
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
			Column: []square.WhereQueryColumn{
				{
					ColumnName:  "id",
					Operator:    Operator.Eq,
					Placeholder: []any{1},
				},
			},
			Relation: &square.WhereRelationQuery{
				ParentTable:   "video",
				ChildrenTable: "video_actor",
				Columns: []square.RelationColumn{
					{
						Parent: "id",
						This:   "video_id",
					},
				},
				Where: &square.WhereQuery{
					Column: []square.WhereQueryColumn{
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
		Orderby: []square.OrderbyQuery{
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

	err = square.FetchQuery(conn, &structuredQuery, &result)

	if err != nil {
		panic(err)
	}

	fmt.Println(result)
}