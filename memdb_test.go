package memdb

import (
	"memdb/internal/utils"
	"sort"
	"testing"
	"time"
)

func TestExpire1(t *testing.T) {
	var db = New()
	db.Set("a", 1, time.Second)
	db.Set("b", 1, 3*time.Second)
	db.Set("c", 1, 5*time.Second)
	db.Set("d", 1, 7*time.Second)
	db.Set("e", 1, 9*time.Second)

	db.Set("c", "1", time.Second)
	time.Sleep(2 * time.Second)

	var keys = db.Keys()
	sort.Strings(keys)
	if !utils.SameStrings(keys, []string{"b", "d", "e"}) {
		t.Fatal()
	}
}

func TestExpire2(t *testing.T) {
	var db = New()
	db.Set("a", 1, time.Second)
	db.Set("b", 1, 3*time.Second)
	db.Set("c", 1, 5*time.Second)
	db.Set("d", 1, 7*time.Second)
	db.Set("e", 1, 29*time.Second)

	db.Set("c", "1", time.Second)
	time.Sleep(15 * time.Second)

	var keys = db.Keys()
	sort.Strings(keys)
	if !utils.SameStrings(keys, []string{"e"}) {
		t.Fatal()
	}
}
