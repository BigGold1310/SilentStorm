package utils

func ToPtr[Value interface{}](value Value) *Value {
	return &value
}

func StringPtrToString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
