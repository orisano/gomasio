package socketio

import (
	"encoding/json"
	"errors"
)

type Event struct {
	Name string
	Args [][]byte
}

func (e *Event) MarshalJSON() ([]byte, error) {
	msg := make([]json.RawMessage, 0, 1+len(e.Args))
	name, err := json.Marshal(e.Name)
	if err != nil {
		return nil, err
	}
	msg = append(msg, json.RawMessage(name))
	for _, arg := range e.Args {
		msg = append(msg, json.RawMessage(arg))
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
	e.Args = make([][]byte, 0, m-1)
	for _, arg := range msg[1:] {
		e.Args = append(e.Args, arg)
	}
	return nil
}
