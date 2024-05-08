package common

import "fmt"

const query = "(EXISTS " +
	"(SELECT 1 FROM %s AS" +
	" %s_%s" +
	" WHERE" +
	" %s_%s.%s = %s_%s.%s" +
	" AND %s)"

func GenerateWhereQuery(
	childTableName,
	childTableRandomSuffix,
	childTableColumn,
	selfTableName,
	selfTableRandomSuffix,
	selfTableColumn,
	moreQuery string) string {
	return fmt.Sprintf(query,
		childTableName,
		childTableName,
		childTableRandomSuffix,
		childTableName,
		childTableRandomSuffix,
		childTableColumn,
		selfTableName,
		selfTableRandomSuffix,
		selfTableColumn,
		moreQuery)
}
