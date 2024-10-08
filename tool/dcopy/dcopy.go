package dcopy

import (
	"errors"
	"reflect"
	"time"
)

func CopySDT[T any](src T, dest *T) error {
	sv := reflect.ValueOf(src)
	dv := reflect.ValueOf(dest)
	if dv.Kind() != reflect.Pointer || dv.IsNil() {
		return errors.New("nil dest")
	}
	copyRecursive(sv, dv.Elem())
	return nil
}

func CopyT[T any](src T) T {
	return Copy(src).(T)
}

func Copy(src interface{}) interface{} {
	original := reflect.ValueOf(src)
	dest := reflect.New(original.Type()).Elem()
	copyRecursive(original, dest)
	return dest.Interface()
}

func copyRecursive(src, dest reflect.Value) {
	switch src.Kind() {
	case reflect.Ptr:
		original := src.Elem()
		if src.IsNil() || !original.IsValid() {
			dest.Set(reflect.Zero(src.Type()))
			return
		}
		destValue := reflect.New(original.Type())
		copyRecursive(original, destValue.Elem())
		dest.Set(destValue)
	case reflect.Interface:
		if src.IsNil() {
			dest.Set(reflect.Zero(src.Type()))
			return
		}
		original := src.Elem()
		destValue := reflect.New(original.Type()).Elem()
		copyRecursive(original, destValue)
		dest.Set(destValue)
	case reflect.Slice:
		if src.IsNil() {
			dest.Set(reflect.Zero(src.Type()))
			return
		}
		dest.Set(reflect.MakeSlice(src.Type(), src.Len(), src.Cap()))
		for i := 0; i < src.Len(); i++ {
			copyRecursive(src.Index(i), dest.Index(i))
		}
	case reflect.Map:
		if src.IsNil() {
			dest.Set(reflect.Zero(src.Type()))
			return
		}
		dest.Set(reflect.MakeMap(src.Type()))
		for _, key := range src.MapKeys() {
			originValue := src.MapIndex(key)
			destValue := reflect.New(originValue.Type()).Elem()
			copyRecursive(originValue, destValue)
			destKey := Copy(key.Interface())
			dest.SetMapIndex(reflect.ValueOf(destKey), destValue)
		}
	case reflect.Struct:
		t, ok := src.Interface().(time.Time)
		if ok {
			dest.Set(reflect.ValueOf(t))
			return
		} else {
			for i := 0; i < src.NumField(); i++ {
				p := src.Type().Field(i)
				if !p.IsExported() {
					continue
				}
				copyRecursive(src.Field(i), dest.Field(i))
			}
		}
	default:
		dest.Set(src)
	}
}
