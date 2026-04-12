package xss

import (
	"reflect"

	"github.com/microcosm-cc/bluemonday"
)

var sanitizer = bluemonday.UGCPolicy()

func SanitizeStruct(v any) {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Pointer || val.IsNil() {
		return
	}

	sanitizeValue(val.Elem())
}

func sanitizeValue(val reflect.Value) {
	switch val.Kind() {
	case reflect.String:
		if val.CanSet() {
			clean := sanitizer.Sanitize(val.String())
			val.SetString(clean)
		}

	case reflect.Pointer:
		if !val.IsNil() {
			sanitizeValue(val.Elem())
		}

	case reflect.Interface:
		if !val.IsNil() {
			sanitizeValue(val.Elem())
		}

	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			if field.CanSet() {
				sanitizeValue(field)
			}
		}

	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			sanitizeValue(val.Index(i))
		}

	case reflect.Map:
		iter := val.MapRange()
		for iter.Next() {
			key := iter.Key()
			value := iter.Value()

			if key.Kind() == reflect.String {
				clean := sanitizer.Sanitize(key.String())
				val.SetMapIndex(key, reflect.ValueOf(clean))
			}

			if value.Kind() == reflect.String {
				clean := sanitizer.Sanitize(value.String())
				val.SetMapIndex(key, reflect.ValueOf(clean))
			} else {
				sanitizeValue(value)
			}
		}
	}
}
