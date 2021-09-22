package term

import (
	"fmt"
)

type NextWriter struct {
	isClosed bool
	buffer   chan []byte
}

func NewNextWriter() *NextWriter {
	return &NextWriter{
		isClosed: false,
		buffer:   make(chan []byte, 16),
	}
}

func (w *NextWriter) Write(p []byte) (bytes int, err error) {
	// fast path
	if w.isClosed {
		return 0, fmt.Errorf("the nextWriter has been closed.")
	}

	defer func() {
		if recover() != nil {
			bytes = 0
			err = fmt.Errorf("the nextWriter has been closed.")
		}
	}()

	w.buffer <- p

	return len(p), nil
}

func (w *NextWriter) Read() ([]byte, int, error) {
	// fast path
	if w.isClosed {
		return nil, 0, fmt.Errorf("the nextWriter has been closed.")
	}

	b, opened := <-w.buffer
	if !opened {
		return nil, 0, fmt.Errorf("the nextWriter has been closed.")
	}

	return b, len(b), nil
}

func (w *NextWriter) Close() {
	if !w.isClosed {
		w.isClosed = true
		close(w.buffer)
	}
}
