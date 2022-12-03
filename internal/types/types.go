package types

type (
	Element struct {
		Value    interface{}
		ExpireAt int64 // ms, -1 as forever
	}
)
