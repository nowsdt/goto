package utils

func ContainsValue(m map[string]string, value string) bool {
	for _, v := range m {
		if value == v {
			return true
		}
	}

	return false
}
