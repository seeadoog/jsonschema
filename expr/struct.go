package expr

import (
	"fmt"
	"reflect"
)

func getFieldOfStruct(forceType bool, rv reflect.Value, name string) any {
	switch rv.Kind() {
	case reflect.Struct:
		fv := rv.FieldByName(name)
		if !fv.IsValid() {
			return nil
		}
		return structValueToVm(forceType, fv.Interface())
	case reflect.Ptr:
		if !rv.IsNil() {
			return getFieldOfStruct(forceType, rv.Elem(), name)
		}
		return nil
	case reflect.Map:

		kv, ok := structValConvert(nil, rv.Type().Key(), name)
		if ok {
			return nil
		}

		data := rv.MapIndex(kv)
		if !data.IsValid() {
			return nil
		}
		return structValueToVm(forceType, data.Interface())
	default:
		return nil
	}
}

func setValStruct(ctx *Context, fv reflect.Value, val any) {
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
	default:
		v, ok := structValConvert(ctx, fv.Type(), val)
		if ok {
			fv.Set(v)
		} else {
			panic(fmt.Sprintf("cannot set val '%v':%s to type %s", val, reflect.TypeOf(val).String(), fv.Type().String()))
		}
	}
}

func structValConvert(ctx *Context, t reflect.Type, v any) (vv reflect.Value, ok bool) {
	if reflect.TypeOf(v) == t {
		return reflect.ValueOf(v), true
	}
	if t.Kind() == reflect.Interface && v != nil {
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
	if v == nil {
		return tv, true
	}

	switch t.Kind() {
	case reflect.String:
		tv.SetString(StringOf(v))
		return tv, true
		//return reflect.ValueOf(StringOf(v)), true
	case reflect.Int, reflect.Int8, reflect.Int64, reflect.Int32, reflect.Int16:
		tv.SetInt(int64(NumberOf(v)))
		return tv, true
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		tv.SetUint(uint64(NumberOf(v)))
		return tv, true
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

			//if fi.Anonymous{
			//
			//}
			//
			fv := tv.Field(i)
			ftv, ok := structValConvert(ctx, fi.Type, fieldV)
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
		if t.Elem().Kind() == reflect.Uint8 {
			tv.SetBytes(ToBytes(StringOf(v)))
			return tv, true
		}

		obj, ok := v.([]any)
		if !ok {
			return vv, false
		}
		tvp := tv
		for _, a := range obj {
			ftv, ok := structValConvert(ctx, t.Elem(), a)
			if !ok {
				return vv, false
			}
			tvp = reflect.Append(tvp, ftv)
		}
		//tv.Set(tvp)
		return tvp, true
	case reflect.Map:
		obj, ok := v.(map[string]any)
		if !ok {
			return vv, false
		}
		tvp := reflect.MakeMap(t)

		for key, val := range obj {
			mk, ok := structValConvert(ctx, t.Key(), key)
			if !ok {
				return vv, false
			}
			nv, ok := structValConvert(ctx, t.Elem(), val)
			if !ok {
				return vv, false
			}
			tvp.SetMapIndex(mk, nv)
		}
		return tvp, true

	case reflect.Interface:
		return reflect.ValueOf(v), true

	case reflect.Func:
		val, ok := v.(Val)
		if !ok {
			return vv, false
		}
		return reflect.MakeFunc(t, func(args []reflect.Value) []reflect.Value {
			argvs := make([]any, len(args))
			for i, arg := range args {
				argvs[i] = arg.Interface()
			}
			res := RunLambda(ctx, val, argvs...)
			switch t.NumOut() {
			case 0:
				return nil
			case 1:
				data, _ := structValConvert(ctx, t.Out(0), res)
				return []reflect.Value{data}
			default:
				rr := make([]reflect.Value, t.NumOut())
				arr, _ := res.([]any)

				for i := 0; i < t.NumOut(); i++ {
					var vvv any
					if i < len(arr) {
						vvv = arr[i]
					} else {
					}
					data, _ := structValConvert(ctx, t.Out(i), vvv)
					rr[i] = data
				}
				return rr
			}

		}), true
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

func setFieldOfStruct(ctx *Context, rv reflect.Value, name string, val any) {
	switch rv.Kind() {
	case reflect.Struct:
		fv := rv.FieldByName(name)
		if !fv.IsValid() {
			return
		}
		setValStruct(ctx, fv, val)
	case reflect.Ptr:
		if !rv.IsNil() {
			setFieldOfStruct(ctx, rv.Elem(), name, val)
			return
		}
	case reflect.Map:
		v, ok := structValConvert(ctx, rv.Type().Elem(), val)
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

func getIndexOfSlice(ctx *Context, forceType bool, rv reflect.Value, v any) any {
	switch rv.Kind() {
	case reflect.Ptr:
		if !rv.IsNil() {
			return getIndexOfSlice(ctx, forceType, rv.Elem(), v)
		}
		return nil
	case reflect.Slice:
		idx := int(NumberOf(v))
		if idx >= rv.Len() {
			return nil
		}
		return structValueToVm(forceType, rv.Index(idx).Interface())
	case reflect.Map:
		keyType := rv.Type().Key()

		kk, ok := structValConvert(ctx, keyType, v)
		if !ok {
			return nil
		}
		ssv := rv.MapIndex(kk)
		if ssv.IsValid() {
			return ssv.Interface()
		}
		return nil
	default:
		return nil
	}
}

func setIndexOfStruct(ctx *Context, rv reflect.Value, idx int, val any) {
	switch rv.Kind() {
	case reflect.Ptr:
		if !rv.IsNil() {
			setIndexOfStruct(ctx, rv.Elem(), idx, val)
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
		v, ok := structValConvert(ctx, rv.Type().Elem(), val)
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

var (
	contextType = reflect.TypeOf(new(Context))
	optionType  = reflect.TypeOf(new(Options))
)

func callFunc(ctx *Context, fv reflect.Value, args []Val) (ress any) {
	ft := fv.Type()
	fvls := make([]reflect.Value, 0, ft.NumIn())

	offset := 0
	isVariadic := ft.IsVariadic()
	for i := 0; i < ft.NumIn(); i++ {

		argi := ft.In(i)

		if i == 0 && argi == contextType {
			fvls = append(fvls, reflect.ValueOf(ctx))
			offset = 1
			continue
		}

		var argv any

		if isVariadic && i-offset == ft.NumIn()-1 {
			argvs := []any{}
			for j := i - offset; j < len(args); j++ {

				vadv, ok := args[j].(*VariadicVal)
				if ok {
					argvs = append(argvs, vadv.ArrVal(ctx)...)
				} else {
					e := args[j].Val(ctx)
					argvs = append(argvs, e)
				}

			}
			argv = argvs
		} else {
			if i-offset < len(args) {
				argv = args[i-offset].Val(ctx)
			}
		}

		if i == ft.NumIn()-1 && argi == optionType {
			o, _ := argv.(map[string]any)
			fvls = append(fvls, reflect.ValueOf(&Options{data: o}))
			continue
		}

		v, ok := structValConvert(ctx, argi, argv)
		if !ok {
			return newErrorf("faile to call '%s' arg  type is not support: %v", argi.Name(), argi.String())
		}
		fvls = append(fvls, v)
	}
	var res []reflect.Value
	if isVariadic {
		res = fv.CallSlice(fvls)
	} else {
		res = fv.Call(fvls)
	}
	if len(res) == 0 {
		return nil
	}
	if len(res) == 1 {
		return res[0].Interface()
	}
	rrss := make([]any, len(res))
	for i, re := range res {
		rrss[i] = re.Interface()
	}
	return rrss
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
	return callFunc(ctx, fv, args), true
}
