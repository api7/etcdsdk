package etcdsdk

// InArray checks if the given value is in the given array
func InArray[V string | int | HookMethod | ID](key V, array []V) bool {
	for _, item := range array {
		if key == item {
			return true
		}
	}

	return false
}

// Pagination paginates the given list
func Pagination(rows []interface{}, size, number int) []interface{} {
	if size > 0 && number > 0 {
		skip := (number - 1) * size
		if skip > len(rows) {
			return []interface{}{}
		}

		end := skip + size
		if end >= len(rows) {
			return rows[skip:]
		}

		return rows[skip:end]
	}

	return rows
}
