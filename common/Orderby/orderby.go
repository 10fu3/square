package Orderby

type OrderBy string

func (o OrderBy) String() string {
	return string(o)
}

const (
	Asc            OrderBy = "asc"
	AscNullsFirst  OrderBy = "asc_nulls_first"
	Desc           OrderBy = "desc"
	DescNullsFirst OrderBy = "desc_nulls_first"
)
