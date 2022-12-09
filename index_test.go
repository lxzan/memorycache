package memorycache

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	var as = assert.New(t)

	t.Run("", func(t *testing.T) {
		var db = New(WithTTLCheckInterval(100 * time.Millisecond))
		db.Set("a", 1, time.Second)
		db.Set("b", 1, 3*time.Second)
		db.Set("c", 1, 5*time.Second)
		db.Set("d", 1, 7*time.Second)
		db.Set("e", 1, 9*time.Second)
		db.Set("c", 1, time.Second)

		time.Sleep(2 * time.Second)
		as.ElementsMatch(db.Keys(), []string{"b", "d", "e"})
	})

	t.Run("", func(t *testing.T) {
		var db = New(WithTTLCheckInterval(100 * time.Millisecond))
		db.Set("a", 1, time.Second)
		db.Set("b", 1, 2*time.Second)
		db.Set("c", 1, 5*time.Second)
		db.Set("d", 1, 7*time.Second)
		db.Set("e", 1, 29*time.Second)
		db.Set("a", 1, 4*time.Second)

		time.Sleep(3 * time.Second)
		as.ElementsMatch(db.Keys(), []string{"a", "c", "d", "e"})
	})

	t.Run("", func(t *testing.T) {
		var db = New(WithTTLCheckInterval(100 * time.Millisecond))
		db.Set("a", 1, time.Second)
		db.Set("b", 1, 2*time.Second)
		db.Set("c", 1, 4*time.Second)
		db.Set("d", 1, 7*time.Second)
		db.Set("d", 1, 4*time.Second)

		time.Sleep(5 * time.Second)
		as.Equal(0, db.Len())
	})
}
