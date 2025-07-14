package utility

import (
	"fmt"
	"reflect"
	"strconv"
)

// GetFieldIDTag 获取结构体的ID字段
func GetFieldIDTag(obj any) string {
	val := reflect.ValueOf(obj)
	stype := reflect.TypeOf(obj)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		stype = stype.Elem()
	}

	for i := 0; i < stype.NumField(); i++ {
		fieldType := stype.Field(i)
		id_property := fieldType.Tag.Get("id_property")
		isID, err := strconv.ParseBool(id_property)
		if err != nil {
			continue
		}
		if isID {
			rst := val.Field(i).Interface()
			return fmt.Sprintf("%v", rst)
		}
	}
	return ""
}
