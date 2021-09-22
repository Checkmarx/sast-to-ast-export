package sliceutils

func Contains(needle interface{}, data []interface{}) bool {
	for _, v := range data {
		if needle == v {
			return true
		}
	}
	return false
}

func Unique(data []interface{}) []interface{} {
	var out []interface{}
	for _, v := range data {
		if Contains(v, out) {
			continue
		}
		out = append(out, v)
	}
	return out
}

func ConvertStringToInterface(data []string) []interface{} {
	out := make([]interface{}, len(data))
	for i, v := range data {
		out[i] = v
	}
	return out
}

func ConvertInterfaceToString(data []interface{}) []string {
	out := make([]string, len(data))
	for i, v := range data {
		out[i] = v.(string)
	}
	return out
}
