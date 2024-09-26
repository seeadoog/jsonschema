package jsonschema

func typeOf(v interface{}) _type {
	switch v.(type) {
	case string:
		return typeString
	case []interface{}:
		return typeArray
	case map[string]interface{}:
		return typeObject
	case float64:
		return typeNumber
	case bool:
		return typeBool
	case nil:
		return typeKnown
	default:
		return typeKnown
	}
}
