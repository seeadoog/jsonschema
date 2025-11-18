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
	if reflect.TypeOf(v) == t {
		return reflect.ValueOf(v), true
	}
	if t.Kind() == reflect.Interface {
		return reflect.ValueOf(v), true
	}
	var tv reflect.Value
	isPtr := false
	if t.Kind() == reflect.Ptr {
		tv = reflect.New(t.Elem()).Elem()
		t = t.Elem()
		isPtr = true
	} else {
		tv = reflect.New(t).Elem()
	}

	switch t.Kind() {
	case reflect.String:
		tv.SetString(StringOf(v))
		return tv, true
		//return reflect.ValueOf(StringOf(v)), true
	case reflect.Int, reflect.Int8, reflect.Int64, reflect.Int32, reflect.Int16:
		tv.SetInt(int64(NumberOf(v)))
		return tv, true
		//return reflect.ValueOf(int(NumberOf(v))), true
	//case reflect.Int8:
	//	tv.SetInt(int64(NumberOf(v)))
	//	return tv.Elem(), true
	//	//return reflect.ValueOf(int8(NumberOf(v))), true
	//case reflect.Int16:
	//	return reflect.ValueOf(int16(NumberOf(v))), true
	//case reflect.Int32:
	//	return reflect.ValueOf(int32(NumberOf(v))), true
	//case reflect.Int64:
	//	return reflect.ValueOf(int64(NumberOf(v))), true
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		tv.SetUint(uint64(NumberOf(v)))
		return tv, true
		//return reflect.ValueOf(uint(NumberOf(v))), true
	//case reflect.Uint8:
	//	return reflect.ValueOf(uint8(NumberOf(v))), true
	//case reflect.Uint16:
	//	return reflect.ValueOf(uint16(NumberOf(v))), true
	//case reflect.Uint32:
	//	return reflect.ValueOf(uint32(NumberOf(v))), true
	//case reflect.Uint64:
	//	return reflect.ValueOf(uint64(NumberOf(v))), true
	case reflect.Float32, reflect.Float64:
		tv.SetFloat(NumberOf(v))
		return tv, true
		//return reflect.ValueOf(float32(NumberOf(v))), true
	//case reflect.Float64:
	//	return reflect.ValueOf(float64(NumberOf(v))), true
	case reflect.Bool:
		tv.SetBool(BoolOf(v))
		return tv, true
	//return reflect.ValueOf(BoolOf(v)), true
	case reflect.Struct:
		obj, ok := v.(map[string]any)
		if !ok {
			return vv, false
		}
		for i := 0; i < t.NumField(); i++ {
			fi := t.Field(i)

			fieldV := obj[fi.Name]
			if fieldV == nil {
				continue
			}

			fv := tv.Field(i)
			ftv, ok := structValConvert(fi.Type, fieldV)
			if !ok {
				return vv, false
			}
			fv.Set(ftv)
		}
		if isPtr {
			return tv.Addr(), true
		}
		return tv, true
	case reflect.Slice:
		obj, ok := v.([]any)
		if !ok {
			return vv, false
		}
		tvp := tv
		for _, a := range obj {
			ftv, ok := structValConvert(t.Elem(), a)
			if !ok {
				return vv, false
			}
			tvp = reflect.Append(tvp, ftv)
		}
		tv.Set(tvp)
		return tv, true
	case reflect.Interface:
		return reflect.ValueOf(v), true
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
			//rvc := rv
			//var e reflect.Value
			//for i := 0; i <= idx-rv.Len(); i++ {
			//	e = reflect.New(rv.Type().Elem()).Elem()
			//	rvc = reflect.Append(rvc, e)
			//}
			//v, ok := structValConvert(rv.Type().Elem(), val)
			//if ok {
			//	e.Set(v)
			//	return
			//}
			//
			//rv.Set(rvc)

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

func callFuncByReflect(ctx *Context, f *objFuncVal, v any, args []Val) (ress any, ok bool) {

	if v == nil {
		return nil, false
	}
	vv := reflect.ValueOf(v)
	fv := vv.MethodByName(f.funcName)
	if !fv.IsValid() {
		return nil, false
	}
	ft := fv.Type()
	fvls := make([]reflect.Value, ft.NumIn())

	if len(args) != ft.NumIn() {
		return newErrorf("faile to call '%s' arg num not match,want %d, got %d", f.funcName, ft.NumIn(), len(args)), true
	}
	for i := 0; i < ft.NumIn(); i++ {

		argi := ft.In(i)

		v, ok := structValConvert(argi, args[i].Val(ctx))
		if !ok {
			return newErrorf("faile to call '%s' arg  type is not support: %v", argi.Name(), argi.String()), true
		}
		fvls[i] = v
	}
	res := fv.Call(fvls)
	if len(res) == 0 {
		return nil, true
	}
	return structValueToVm(false, res[0].Interface()), true
}
