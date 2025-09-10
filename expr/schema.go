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
	RegisterFunc("error", func(ctx *Context, args ...Val) any {
		if len(args) == 0 {
			return nil
		}
		return args[0].Val(ctx)
	})
	jsonschema.RegisterValidator("script", NewScript)

	//RegisterExp("throw", func(o map[string]any, val any) (Expr, error) {
	//})
}

func (s *ScriptSchema) Validate(c *jsonschema.ValidateCtx, value interface{}) {
	//v, ok := value.(map[string]any)
	//if !ok {
	//	return
	//}
	ctx := &Context{
		table: map[string]any{
			"$": value,
		},
	}
	err := ctx.Exec(s.expr)
	er, ok := err.(*throwError)
	if ok {
		c.AddErrorInfo(s.path, fmt.Sprintf("%v", er.data))
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
