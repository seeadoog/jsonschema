package jsonschema

import "github.com/seeadoog/jsonschema/v2/utils"

// UnmarshalFromMap 将map 中的值序列化到 struct 中
func UnmarshalFromMap(in interface{}, template interface{}) error {
	return utils.UnmarshalFromMap(in, template)
	//v := reflect.ValueOf(template)
	//if v.Kind() != reflect.Ptr || v.IsNil() {
	//	panic("template value is nil or not pointer")
	//}
	//return unmarshalObject2Struct("", in, v)
}
