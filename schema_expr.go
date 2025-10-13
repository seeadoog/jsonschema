package jsonschema

import (
	"fmt"
	expr2 "github.com/seeadoog/jsonschema/v2/expr"
)

type ScriptSchema struct {
	expr expr2.Expr
	path string
}

func init() {
	//RegisterFunc("error", func(ctx *Context, args ...Val) any {
	//	if len(args) == 0 {
	//		return nil
	//	}
	//	return args[0].Val(ctx)
	//})
	RegisterValidator("script", NewScript)

	//RegisterExp("throw", func(o map[string]any, val any) (Expr, error) {
	//})
}

func (s *ScriptSchema) Validate(c *ValidateCtx, value interface{}) {
	//v, ok := value.(map[string]any)
	//if !ok {
	//	return
	//}
	ctx := expr2.NewContext(map[string]any{
		"$": value,
	})
	ctx.Exec(s.expr)
	ret := ctx.GetReturn()
	if len(ret) > 0 {
		c.AddErrorInfo(s.path, fmt.Sprintf("err :%v", ret))
	}

}

var NewScript NewValidatorFunc = func(i interface{}, path string, parent Validator) (Validator, error) {

	e, err := expr2.ParseFromJSONObj(i)
	if err != nil {
		return nil, fmt.Errorf("parse as script: %w %s", err, path)
	}
	return &ScriptSchema{
		expr: e,
		path: path,
	}, nil
}

type throwError struct {
	data any
}

func (t *throwError) Error() string {
	return fmt.Sprintf("throw error, %s", t.data)
}

type throwExp struct {
	val expr2.Val
}

func (t *throwExp) Exec(c *expr2.Context) error {
	return &throwError{t.val.Val(c)}
}
