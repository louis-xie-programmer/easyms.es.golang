package utility

func NullArrayIntCheck(val []int32) bool {
	if val == nil || len(val) == 0 {
		return true
	}
	return false
}

func NullMapStringCheck(val map[string]string) bool {
	if val == nil || len(val) == 0 {
		return true
	}
	return false
}
