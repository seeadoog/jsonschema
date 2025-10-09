package expr

//type Object interface {
//	Get(key string) any
//	Set(key string, value any)
//	Range(f func(key string, value any) bool)
//}
//
//type Slice interface {
//	Get(key int) any
//	Set(key int, value any)
//	Range(f func(key int, value any) bool)
//}
//
//type mapObject map[string]any
//
//func (m mapObject) Get(key string) any {
//	return m[key]
//}
//func (m mapObject) Set(key string, value any) {
//	m[key] = value
//}
//func (m mapObject) Range(f func(key string, value any) bool) {
//	for k, v := range m {
//		if !f(k, v) {
//			return
//		}
//	}
//}
