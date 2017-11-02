package socketio

import (
	"encoding/json"
	"errors"
)

type Event struct {
	Name string
	Args []byte
}

func (e *Event) MarshalJSON() ([]byte, error) {
	name, err := json.Marshal(e.Name)
	if err != nil {
		return nil, err
	}
	msg := []json.RawMessage{json.RawMessage(name)}
	if len(e.Args) > 0 {
		msg = append(msg, json.RawMessage(e.Args))
	}
	return json.Marshal(msg)
}

func (e *Event) UnmarshalJSON(p []byte) error {
	var msg []json.RawMessage
	if err := json.Unmarshal(p, &msg); err != nil {
		return err
	}
	m := len(msg)
	if m == 0 {
		return errors.New("invalid length event")
	}
	if err := json.Unmarshal(msg[0], &e.Name); err != nil {
		return err
	}
	if m > 1 {
		e.Args = msg[1]
	}
	return nil
}
