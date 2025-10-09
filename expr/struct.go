package expr

import "reflect"

func getFieldOfStruct(rv reflect.Value, name string) any {
	switch rv.Kind() {
	case reflect.Struct:
		fv := rv.FieldByName(name)
		if !fv.IsValid() {
			return nil
		}
		return structValueToVm(true, fv.Interface())
	case reflect.Ptr:
		if !rv.IsNil() {
			return getFieldOfStruct(rv.Elem(), name)
		}
		return nil
	case reflect.Map:
		return structValueToVm(true, rv.MapIndex(reflect.ValueOf(name)).Interface())
	default:
		return nil
	}
}

func setValStruct(fv reflect.Value, val any) {
	switch fv.Kind() {
	case reflect.String:
		fv.SetString(StringOf(val))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fv.SetInt(int64(NumberOf(val)))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fv.SetUint(uint64(NumberOf(val)))
	case reflect.Float32, reflect.Float64:
		fv.SetFloat(NumberOf(val))
	case reflect.Bool:
		fv.SetBool(BoolOf(val))
	}
}

func structValConvert(t reflect.Type, v any) (vv reflect.Value, ok bool) {
	switch t.Kind() {
	case reflect.String:
		return reflect.ValueOf(StringOf(v)), true
	case reflect.Int:
		return reflect.ValueOf(int(NumberOf(v))), true
	case reflect.Int8:
		return reflect.ValueOf(int8(NumberOf(v))), true
	case reflect.Int16:
		return reflect.ValueOf(int16(NumberOf(v))), true
	case reflect.Int32:
		return reflect.ValueOf(int32(NumberOf(v))), true
	case reflect.Int64:
		return reflect.ValueOf(int64(NumberOf(v))), true
	case reflect.Uint:
		return reflect.ValueOf(uint(NumberOf(v))), true
	case reflect.Uint8:
		return reflect.ValueOf(uint8(NumberOf(v))), true
	case reflect.Uint16:
		return reflect.ValueOf(uint16(NumberOf(v))), true
	case reflect.Uint32:
		return reflect.ValueOf(uint32(NumberOf(v))), true
	case reflect.Uint64:
		return reflect.ValueOf(uint64(NumberOf(v))), true
	case reflect.Float32:
		return reflect.ValueOf(float32(NumberOf(v))), true
	case reflect.Float64:
		return reflect.ValueOf(float64(NumberOf(v))), true
	case reflect.Bool:
		return reflect.ValueOf(BoolOf(v)), true
	default:
		return vv, false
	}
}

func structValueToVm(force bool, vv any) any {
	if !force {
		return vv
	}
	switch v := vv.(type) {
	case string, float64, bool, []any, map[string]interface{}, []byte:
		return vv
	case int:
		return float64(v)
	case int8:
		return float64(v)
	case int16:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case uint:
		return float64(v)
	case uint8:
		return float64(v)
	case uint16:
		return float64(v)
	case uint32:
		return float64(v)
	case uint64:
		return float64(v)
	case float32:
		return float64(v)
	default:
		return v
	}
}

func setFieldOfStruct(rv reflect.Value, name string, val any) {
	switch rv.Kind() {
	case reflect.Struct:
		fv := rv.FieldByName(name)
		if !fv.IsValid() {
			return
		}
		setValStruct(fv, val)
	case reflect.Ptr:
		if !rv.IsNil() {
			setFieldOfStruct(rv.Elem(), name, val)
			return
		}
	case reflect.Map:
		v, ok := structValConvert(rv.Type().Elem(), val)
		if !ok {
			return
		}
		if rv.Type().Key().Kind() != reflect.String {
			return
		}
		rv.SetMapIndex(reflect.ValueOf(name), v)
	default:
	}
}

func getIndexOfSlice(rv reflect.Value, idx int) any {
	switch rv.Kind() {
	case reflect.Ptr:
		if !rv.IsNil() {
			return getIndexOfSlice(rv.Elem(), idx)
		}
		return nil
	case reflect.Slice:
		if idx >= rv.Len() {
			return nil
		}
		return structValueToVm(true, rv.Index(idx).Interface())
	default:
		return nil
	}
}

func setIndexOfStruct(rv reflect.Value, idx int, val any) {
	switch rv.Kind() {
	case reflect.Ptr:
		if !rv.IsNil() {
			setIndexOfStruct(rv.Elem(), idx, val)
		}
	case reflect.Slice:
		if idx >= rv.Len() {
			return
		}
		v, ok := structValConvert(rv.Type().Elem(), val)
		if !ok {
			return
		}
		rv.Index(idx).Set(v)

	default:

	}
}

func lenOfStruct(rv reflect.Value) int64 {
	switch rv.Kind() {
	case reflect.Ptr:
		if !rv.IsNil() {
			return lenOfStruct(rv.Elem())
		}
	case reflect.Slice:
		return int64(rv.Len())
	}
	return 0
}
