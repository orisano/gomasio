package socketio

import (
	"encoding/json"
	"testing"
)

func TestEvent_MarshalJSON(t *testing.T) {
	ts := []struct {
		event    Event
		expected string
	}{
		{
			event:    Event{Name: "hello"},
			expected: `["hello"]`,
		},
		{
			event:    Event{Name: "hello", Args: []byte(`{"hello":"world"}`)},
			expected: `["hello",{"hello":"world"}]`,
		},
	}

	for _, tc := range ts {
		b, err := json.Marshal(&tc.event)
		if err != nil {
			t.Error(err)
			continue
		}
		if got := string(b); got != tc.expected {
			t.Errorf("unexpected json. expected: %v, but got: %v", tc.expected, got)
		}
	}
}
