package common

func IfThen(condition bool, trueVal interface{}, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	} else {
		return falseVal
	}
}
