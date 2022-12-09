package utils

import (
	"sort"
)

type Integer interface {
	int | int64 | int32 | uint | uint64 | uint32
}

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

func ToBinaryNumber[T Integer](n T) T {
	var x T = 1
	for x < n {
		x *= 2
	}
	return x
}

func Uniq[T comparable](arr []T) []T {
	var m = make(map[T]struct{}, len(arr))
	var list = make([]T, 0, len(arr))
	for _, item := range arr {
		m[item] = struct{}{}
	}
	for k, _ := range m {
		list = append(list, k)
	}
	return list
}
