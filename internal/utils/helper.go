package utils

import "sort"

func SameStrings(arr1, arr2 []string) bool {
	sort.Strings(arr1)
	sort.Strings(arr2)
	var n = len(arr1)
	if n != len(arr2) {
		return false
	}
	for i := 0; i < n; i++ {
		if arr1[i] != arr2[i] {
			return false
		}
	}
	return true
}
