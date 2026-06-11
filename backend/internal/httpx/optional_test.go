package httpx

import (
	"encoding/json"
	"testing"
	"time"
)

// TestOptionalPatchSemantics locks in the three-state PATCH behaviour that the
// task update endpoint depends on: absent vs explicit null vs a real value.
func TestOptionalPatchSemantics(t *testing.T) {
	type patch struct {
		DueDate Optional[time.Time] `json:"dueDate"`
	}

	t.Run("absent leaves field untouched", func(t *testing.T) {
		var p patch
		mustUnmarshal(t, `{}`, &p)
		if p.DueDate.Present {
			t.Error("absent field should not be Present")
		}
	})

	t.Run("explicit null clears the field", func(t *testing.T) {
		var p patch
		mustUnmarshal(t, `{"dueDate": null}`, &p)
		if !p.DueDate.Present || !p.DueDate.Null {
			t.Errorf("null should be Present and Null, got %+v", p.DueDate)
		}
	})

	t.Run("value sets the field", func(t *testing.T) {
		var p patch
		mustUnmarshal(t, `{"dueDate": "2026-01-02T15:04:05Z"}`, &p)
		if !p.DueDate.Present || p.DueDate.Null {
			t.Fatalf("value should be Present and not Null, got %+v", p.DueDate)
		}
		want := time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC)
		if !p.DueDate.Value.Equal(want) {
			t.Errorf("got %v, want %v", p.DueDate.Value, want)
		}
	})
}

func mustUnmarshal(t *testing.T, data string, v any) {
	t.Helper()
	if err := json.Unmarshal([]byte(data), v); err != nil {
		t.Fatalf("unmarshal %q: %v", data, err)
	}
}
