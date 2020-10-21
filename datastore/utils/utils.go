package utils

// Temporary function for Postgres inserts formatting
func ToVarcharArray(elems []string) string {
	var result string

	last := len(elems) - 1
	for i := range elems {
		result += `"` + elems[i] + `"`

		if i != last {
			result += `, `
		}
	}

	return `{` + result + `}`
}
