package utils

func StringSliceToMap(values []string) map[string]bool {
	result := make(map[string]bool)
	for _, val := range values {
		result[val] = true
	}
	return result
}
