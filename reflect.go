package karma

import (
	"fmt"
	"reflect"
	"strconv"
)

func DescribeDeep(prefixKey string, obj interface{}) *Context {
	ctx := &Context{}
	describeDeep(ctx, obj, prefixKey, "")
	return ctx
}

func describeDeep(ctx *Context, obj interface{}, prefix string, key string) {
	resource := reflect.Indirect(reflect.ValueOf(obj))

	for resource.Kind() == reflect.Ptr {
		resource = resource.Elem()
	}

	prefixKey := joinPrefixKey(prefix, key)

	resourceType := resource.Type()
	switch resource.Kind() {
	case reflect.Struct:
		for index := 0; index < resourceType.NumField(); index++ {
			resourceField := resource.Field(index)
			if !resourceField.CanInterface() {
				continue
			}
			structField := resourceType.Field(index)
			fieldName := string(structField.Name)
			describeDeep(ctx, resourceField.Interface(), prefixKey, fieldName)
		}
	case reflect.Slice:
		for i := 0; i < resource.Len(); i++ {
			field := reflect.Indirect(resource.Index(i))
			if !field.CanInterface() {
				continue
			}
			describeDeep(ctx, field.Interface(), prefixKey, "["+strconv.Itoa(i)+"]")
		}

	default:
		*ctx = *ctx.Describe(prefixKey, fmt.Sprint(obj))
	}

}

func joinPrefixKey(prefix string, key string) string {
	result := prefix
	if key != "" {
		if key[0] == '[' {
			result = result + key
		} else {
			result = result + "." + key
		}
	}

	return result
}
