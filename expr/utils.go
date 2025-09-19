package expr

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strconv"
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
