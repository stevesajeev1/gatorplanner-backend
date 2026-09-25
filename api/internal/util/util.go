package util

func AnyStringSlice[T ~string](slice []T) []any {
	result := make([]any, len(slice))
	for i, v := range slice {
		result[i] = string(v)
	}
	return result
}
