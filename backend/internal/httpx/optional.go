package httpx

import "encoding/json"

// Optional distinguishes the three JSON states a PATCH field can be in:
//
//	field absent        -> Present == false            (leave unchanged)
//	"field": null       -> Present == true, Null==true (clear the value)
//	"field": <value>    -> Present == true, Value set  (set the value)
//
// This is what makes PATCH semantics correct: omitting due_date keeps it,
// while sending null explicitly clears it.
type Optional[T any] struct {
	Present bool
	Null    bool
	Value   T
}

func (o *Optional[T]) UnmarshalJSON(b []byte) error {
	o.Present = true
	if string(b) == "null" {
		o.Null = true
		return nil
	}
	return json.Unmarshal(b, &o.Value)
}
