package gomasio

type prefixWriter struct {
	wf     WriteFlusher
	prefix []byte
	init   bool
}

func (w *prefixWriter) Write(p []byte) (n int, err error) {
	n = 0
	if w.init {
		x, err := w.wf.Write(w.prefix)
		if err != nil {
			return 0, err
		}
		n += x
		w.init = false
	}
	x, err := w.wf.Write(p)
	if err != nil {
		return n, err
	}
	n += x
	return n, err
}

func (w *prefixWriter) Flush() error {
	return w.wf.Flush()
}

func NewPrefixWriter(wf WriteFlusher, prefix []byte) WriteFlusher {
	return &prefixWriter{
		wf:     wf,
		prefix: prefix,
		init:   true,
	}
}
