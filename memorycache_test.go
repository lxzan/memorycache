package memorycache

import (
	"github.com/lxzan/memorycache/internal/utils"
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
	var cfg = Config{TTLCheckInterval: 2}
	var db = New(cfg)
	db.Set("a", 1, time.Second)
	db.Set("b", 1, 3*time.Second)
	db.Set("c", 1, 5*time.Second)
	db.Set("d", 1, 7*time.Second)
	db.Set("e", 1, 29*time.Second)

	db.Set("c", "1", time.Second)
	time.Sleep(3 * time.Second)

	var keys = db.Keys()
	sort.Strings(keys)
	if !utils.SameStrings(keys, []string{"d", "e"}) {
		t.Fatal()
	}
}
