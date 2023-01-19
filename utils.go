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
func Pagination(rows []interface{}, pageSize, pageNumber int) []interface{} {
	if pageSize > 0 && pageNumber > 0 {
		skip := (pageNumber - 1) * pageSize
		if skip > len(rows) {
			return []interface{}{}
		}

		end := skip + pageSize
		if end >= len(rows) {
			return rows[skip:]
		}

		return rows[skip:end]
	}

	return rows
}
