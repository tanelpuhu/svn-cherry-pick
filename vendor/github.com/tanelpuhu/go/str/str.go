package str

// InSlice ...
func InSlice(haystack []string, needle string) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}

// IsSeven ...
func IsSeven(val string) bool {
	return "7" == val
}
