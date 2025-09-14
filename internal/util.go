package internal

import "strings"

func PrettyLogStr(field string) string {
	return strings.Replace(field, "\"", "`", -1)
}
