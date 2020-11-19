package socketio

import (
	"encoding/json"
	"fmt"
)

type Event struct {
	Name string
	Args []json.RawMessage
}

func (e *Event) MarshalJSON() ([]byte, error) {
	msg := make([]json.RawMessage, 1+len(e.Args))
	name, err := json.Marshal(e.Name)
	if err != nil {
		return nil, err
	}
	msg[0] = name
	copy(msg[1:], e.Args)
	return json.Marshal(msg)
}

func (e *Event) UnmarshalJSON(p []byte) error {
	var msg []json.RawMessage
	if err := json.Unmarshal(p, &msg); err != nil {
		return err
	}
	m := len(msg)
	if m == 0 {
		return fmt.Errorf("invalid length event")
	}
	if err := json.Unmarshal(msg[0], &e.Name); err != nil {
		return err
	}
	if m > 1 {
		e.Args = make([]json.RawMessage, m-1)
		copy(e.Args, msg[1:])
	}
	return nil
}
