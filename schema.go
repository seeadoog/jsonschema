package jsonschema

import (
	"encoding/json"
	"errors"
	"strings"
)

type Schema struct {
	prop Validator
	i interface{}
}

func (s *Schema)UnmarshalJSON(b []byte)error{
	var i interface{}
	if err:=json.Unmarshal(b,&i);err != nil{
		return err
	}
	s.i = i
	p ,err := NewProp(i,"$")
	if err != nil{
		return err
	}
	s.prop = p
	return nil
}

func (s *Schema)MarshalJSON()(b []byte,err error){
	data,err:=json.Marshal(s.i)
	if err != nil{
		return nil,err
	}
	return data,nil

}

func (s *Schema)Validate(i interface{})error{
	c:=&ValidateCtx{}
	s.prop.Validate(c,i)
	if len(c.errors) == 0{
		return nil
	}
	return errors.New(errsToString(c.errors))
}

func (s *Schema)ValidateError(i interface{})[]Error{
	c:=&ValidateCtx{}
	s.prop.Validate(c,i)
	return c.errors
}


func errsToString(errs []Error)string{
	sb:=strings.Builder{}
	n:=0
	for _, err := range errs {
		n+= len(err.Path)+ len(err.Info)+5
	}
	sb.Grow(n)
	for _, err := range errs {
		sb.WriteString(appendString("'",err.Path,"' ",err.Info,"; "))
	}
	return sb.String()
}