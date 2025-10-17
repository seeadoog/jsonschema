package expr

import (
	"encoding/base64"
	"fmt"
	xxhash "github.com/cespare/xxhash/v2"
	"hash/crc64"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

func BoolOf(v interface{}) bool {
	switch vv := v.(type) {
	case bool:
		return vv
	case float64:
		return vv != 0
	case int:
		return vv != 0
	case string:
		switch vv {
		case "false", "0", "", "False", "FALSE":
			return false
		}
		return true
	case nil:
		return false

	default:
	}
	return v != nil
}

func BoolCond(v interface{}) bool {
	switch vv := v.(type) {
	case bool:
		return vv
	case nil:
		return false
	default:
		return true
	}
}

func StringOf(v interface{}) string {
	switch vv := v.(type) {
	case string:
		return vv
	case bool:
		if vv {
			return "true"
		}
		return "false"
	case float64:
		return strconv.FormatFloat(vv, 'f', -1, 64)
	case int:
		return strconv.Itoa(vv)
	case nil:
		return ""
	case []byte:
		return unsafe.String(unsafe.SliceData(vv), len(vv))
	case *strings.Builder:
		return vv.String()
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.String {
		return rv.String()
	}
	return fmt.Sprintf("%v", v)
}

func NumberOf(v interface{}) float64 {
	switch vv := v.(type) {
	case float64:
		return vv
	case int:
		return float64(vv)
	case bool:
		if vv {
			return 1
		}
		return 0
	case string:
		i, err := strconv.ParseFloat(vv, 64)
		if err != nil {
			return i
		}
		if vv == "true" {
			return 1
		}
		return 0
	}
	return 0
}

func BytesOf(v interface{}) []byte {
	switch vv := v.(type) {
	case []byte:
		return vv
	case string:
		return ToBytes(vv)
	default:
		return nil
	}
}

func ToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func ToString(b []byte) string {
	//return *(*string)(unsafe.Pointer(&b))
	return unsafe.String(unsafe.SliceData(b), len(b))
}

var (
	base64Enc = base64.StdEncoding
)

func base64EncodeToString(src []byte) string {
	buf := make([]byte, base64Enc.EncodedLen(len(src)))
	base64Enc.Encode(buf, src)
	return ToString(buf)
}

func base64DecodeString(s string) ([]byte, error) {
	dbuf := make([]byte, base64Enc.DecodedLen(len(s)))
	n, err := base64Enc.Decode(dbuf, []byte(s))
	return dbuf[:n], err
}

func indexerOf(v any) func(k string) any {
	switch vv := v.(type) {
	case map[string]interface{}:
		return func(k string) any {
			return vv[k]
		}
	case map[string]string:
		return func(k string) any {
			return vv[k]
		}
	default:
		return nil
	}
}

type Options struct {
	data map[string]interface{}
}

func newOption(data map[string]any) *Options {

	return &Options{data: data}
}

func (o *Options) Has(key string) bool {
	if o == nil {
		return false
	}
	_, ok := o.data[key]
	return ok
}

func (o *Options) Get(key string) any {
	if o == nil {
		return nil
	}
	return o.data[key]
}

func (o *Options) GetString(key string) string {
	return StringOf(o.Get(key))
}
func (o *Options) GetStringDef(key string, def string) string {
	v := o.Get(key)
	if v == nil {
		return def
	}
	return StringOf(v)
}

func (o *Options) GetNumber(key string) float64 {
	return NumberOf(o.Get(key))
}

func (o *Options) GetNumberDef(key string, def float64) float64 {
	v := o.Get(key)
	if v == nil {
		return def
	}
	return NumberOf(v)
}

func (o *Options) GetBool(key string) bool {
	return BoolOf(o.Get(key))
}

func (o *Options) Range(f func(k string, v any) bool) {
	if o == nil {
		return
	}
	for k, v := range o.data {
		if !f(k, v) {
			return
		}
	}
}

func (o *Options) RangeKey(key string, f func(k string, v any) bool) {
	m, ok := o.Get(key).(map[string]any)
	if ok {
		for k, v := range m {
			if !f(k, v) {
				return
			}
		}
	}
}

var (
	table = crc64.MakeTable(crc64.ECMA)
)

func calcHash(s string) uint64 {
	//return crc64.Checksum([]byte(s), table)
	return xxhash.Sum64([]byte(s))
}

func hashType(v interface{}) uint64 {
	return 1
}

func HashString(s string) uint64 {
	return xxhash.Sum64([]byte(s))
}
