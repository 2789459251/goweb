package mylog

import (
	"fmt"
	"reflect"
	"strings"
)

func JoinStrings(data ...any) string {

	var sb strings.Builder
	for _, v := range data {
		sb.WriteString(check(v))
	}
	return sb.String()
}

func check(v any) string {
	Value := reflect.ValueOf(v)
	switch Value.Kind() {
	case reflect.String:
		return Value.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
