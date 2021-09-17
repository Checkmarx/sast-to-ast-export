package sliceutils

func Contains(needle interface{}, data []interface{}) bool {
	for _, v := range data {
		if needle == v {
			return true
		}
	}
	return false
}

func ConvertStringToInterface(data []string) []interface{} {
	out := make([]interface{}, len(data))
	for i, v := range data {
		out[i] = v
	}
	return out
}
