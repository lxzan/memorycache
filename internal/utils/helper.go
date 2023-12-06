package utils

type Integer interface {
	int | int64 | int32 | uint | uint64 | uint32
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

func SelectValue[T any](ok bool, a, b T) T {
	if ok {
		return a
	}
	return b
}

func IsSameSlice[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
