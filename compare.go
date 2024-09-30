package jsonschema

import (
	"fmt"
	"reflect"
)

type Compare[A any, W any] struct {
	cmps   sliceMap[*JsonPathCompiled, W]
	fun    func(actual A, def W, ctx Context) bool
	path   string
	info   string
	isInIf bool
}

func (c *Compare[A, W]) Validate(ctx *ValidateCtx, val any) {

	cc, ok := val.(map[string]any)
	if !ok {
		return
	}
	c.cmps.Range(func(jp *JsonPathCompiled, v W) bool {
		data, _ := jp.Get(val)

		ad, ok := data.(A)

		if !ok || !c.fun(ad, v, cc) {
			if c.isInIf {
				ctx.AddError(Error{})
			} else {
				ctx.AddError(Error{
					Path: c.path + "." + jp.rawPath,
					Info: c.info + StringOf(v),
				})
			}

		}
		return true
	})
	//for jp, v := range c.cmps {
	//
	//}
}

type options[T any] struct {
	parseW func(any) (res T, err error)
}

type opt[T any] func(o *options[T])

func withOptParse[T any](pf func(any) (res T, err error)) opt[T] {
	return func(o *options[T]) {
		o.parseW = pf
	}
}

func NewCompareVal[A, W any](fun func(actual A, def W, c Context) bool, info string) NewValidatorFunc {
	return NewCompare(func(actual A, def Value, c Context) bool {
		dv, ok := def.Get(c).(W)
		if !ok {
			return false
		}
		return fun(actual, dv, c)
	}, info, withOptParse(func(a any) (res Value, err error) {
		return parseValue(a)
	}))
}

func NewCompare[A, W any](fun func(actual A, def W, c Context) bool, info string, opts ...opt[W]) NewValidatorFunc {
	return func(i interface{}, path string, parent Validator) (Validator, error) {
		m, ok := i.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("%s is not a map", path)
		}

		opt := &options[W]{}
		for _, o := range opts {
			o(opt)
		}
		if opt.parseW == nil {
			opt.parseW = func(v any) (res W, err error) {
				vv, ok := v.(W)
				if !ok {
					return res, fmt.Errorf("not type %v", reflect.TypeOf(new(W)).Elem())
				}
				return vv, nil
			}
		}
		cvs := sliceMap[*JsonPathCompiled, W]{}
		for key, val := range m {
			jp, err := parseJpathCompiled(key)
			if err != nil {
				return nil, fmt.Errorf("%s.%s is not a valid jsonpath", path, key)
			}

			vv, err := opt.parseW(val)
			if err != nil {
				return nil, fmt.Errorf("%s.%s %w", path, key, err)
			}
			//cvs[jp] = vv
			cvs.Set(jp, vv)
		}
		return &Compare[A, W]{
			cmps:   cvs,
			fun:    fun,
			path:   path,
			info:   info,
			isInIf: isInIf(parent),
		}, nil
	}
}

type CompareSingle[A any, W any] struct {
	path string
	info string
	val  W
	fun  func(actual A, def W) bool
}

func (c *CompareSingle[A, W]) Validate(ctx *ValidateCtx, val any) {
	vv, ok := val.(A)
	if !ok {
		return
	}
	if !c.fun(vv, c.val) {
		ctx.AddError(Error{
			Path: c.path,
			Info: c.info + StringOf(c.val),
		})
	}
}

func NewCompareSingle[A, W any](cmp func(actual A, def W) bool, info string) NewValidatorFunc {
	return func(i interface{}, path string, parent Validator) (Validator, error) {

		v, ok := i.(W)
		if !ok {
			return nil, fmt.Errorf("%s is not %v", path, reflect.TypeOf(new(W)).Elem())
		}
		return &CompareSingle[A, W]{
			path: path,
			info: info,
			val:  v,
			fun:  cmp,
		}, nil
	}
}
