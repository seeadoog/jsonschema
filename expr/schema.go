package expr

import (
	"fmt"
	"github.com/seeadoog/jsonschema/v2"
)

type ScriptSchema struct {
	expr Expr
	path string
}

func InitSchema() {
	//RegisterFunc("error", func(ctx *Context, args ...Val) any {
	//	if len(args) == 0 {
	//		return nil
	//	}
	//	return args[0].Val(ctx)
	//})
	jsonschema.RegisterValidator("script", NewScript)

	//RegisterExp("throw", func(o map[string]any, val any) (Expr, error) {
	//})
}

func (s *ScriptSchema) Validate(c *jsonschema.ValidateCtx, value interface{}) {
	//v, ok := value.(map[string]any)
	//if !ok {
	//	return
	//}
	ctx := NewContext(map[string]any{
		"$": value,
	})
	ctx.Exec(s.expr)
	ret := ctx.GetReturn()
	if len(ret) > 0 {
		c.AddErrorInfo(s.path, fmt.Sprintf("err :%v", ret))
	}

}

var NewScript jsonschema.NewValidatorFunc = func(i interface{}, path string, parent jsonschema.Validator) (jsonschema.Validator, error) {

	e, err := ParseFromJSONObj(i)
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
	val Val
}

func (t *throwExp) Exec(c *Context) error {
	return &throwError{t.val.Val(c)}
}
