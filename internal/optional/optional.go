package optional

import "encoding/json"

// Optional distinguishes a json field that was omitted from one that was
// explicitly set to null, which a plain pointer cannot do. Partial updates
// need that difference: an absent field means "leave it alone", while an
// explicit null means "clear it".
type Optional[T any] struct {
	Present bool
	Value   *T
}

// UnmarshalJSON is only invoked when the key exists in the payload, which is
// what makes Present reliable.
func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	o.Present = true

	if string(data) == "null" {
		o.Value = nil
		return nil
	}

	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	o.Value = &value
	return nil
}

// Set reports whether the field was provided with a non-null value.
func (o Optional[T]) Set() bool {
	return o.Present && o.Value != nil
}

// Cleared reports whether the field was explicitly set to null.
func (o Optional[T]) Cleared() bool {
	return o.Present && o.Value == nil
}
