package utils

import (
	"math/rand"
	"sort"
	"time"
)

type RandomString string

var Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

const (
	Alphabet RandomString = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Numeric  RandomString = "0123456789"
)

func (this RandomString) Generate(n int) string {
	var b = make([]byte, n)
	var length = len(this)
	for i := 0; i < n; i++ {
		var idx = Rand.Intn(length)
		b[i] = this[idx]
	}
	return string(b)
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

func Timestamp() int64 {
	return time.Now().UnixNano() / 1000000
}
