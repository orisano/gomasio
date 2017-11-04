package socketio

import (
	"bytes"
	"io"
	"testing"

	"github.com/orisano/gomasio"
)

type nopFlusher struct {
	w io.Writer
}

func (f *nopFlusher) Write(p []byte) (n int, err error) {
	return f.w.Write(p)
}

func (*nopFlusher) Flush() error {
	return nil
}

func NopFlusher(w io.Writer) gomasio.WriteFlusher {
	return &nopFlusher{w}
}

type testWriterFactory struct {
	w io.Writer
}

func (f *testWriterFactory) NewWriter() gomasio.WriteFlusher {
	return NopFlusher(f.w)
}

func TestContext_Emit(t *testing.T) {
	ts := []struct {
		event    string
		args     []interface{}
		expected string
	}{
		{
			event:    "hello",
			args:     nil,
			expected: `2["hello"]` + "\n",
		},
		{
			event:    "string",
			args:     []interface{}{"hoge"},
			expected: `2["string","hoge"]` + "\n",
		},
		{
			event:    "number",
			args:     []interface{}{1},
			expected: `2["number",1]` + "\n",
		},
		{
			event: "custom",
			args: []interface{}{&struct {
				Id  int    `json:"id"`
				Msg string `json:"msg"`
			}{
				Id:  15,
				Msg: "hello",
			}},
			expected: `2["custom",{"id":15,"msg":"hello"}]` + "\n",
		},
	}
	for _, tc := range ts {
		var b bytes.Buffer
		ctx, err := NewContext(&testWriterFactory{&b}, &Packet{})
		if err != nil {
			t.Error(err)
			continue
		}
		if err := ctx.Emit(tc.event, tc.args...); err != nil {
			t.Error(err)
			continue
		}
		if got := b.String(); got != tc.expected {
			t.Errorf("unexpected emit event. expected: %v, but got: %v", tc.expected, got)
		}
	}
}
