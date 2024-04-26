package validators

import "slices"

func IsAllStringInArray(a, b []string) bool {
	if len(b) == 0 {
		return false
	}
	for _, v := range b {
		if !slices.Contains(a, v) {
			return false
		}
	}
	return true
}
