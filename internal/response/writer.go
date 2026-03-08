package response

import (
	"fmt"
	"io"

	"github.com/WadeGulbrandsen/httpfromtcp/internal/headers"
)

type writerState int

const (
	writerStateStatusLine writerState = iota
	writerStateHeaders
	writerStateBody
	writerStateTrailers
	writerStateDone
)

type Writer struct {
	writerState writerState
	writer      io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writerState: writerStateStatusLine,
		writer:      w,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerState != writerStateStatusLine {
		return fmt.Errorf("cannot write status line in state %d", w.writerState)
	}
	defer func() { w.writerState = writerStateHeaders }()
	_, err := w.writer.Write(getStatusLine(statusCode))
	return err
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	if w.writerState != writerStateHeaders {
		return fmt.Errorf("cannot write headers in state %d", w.writerState)
	}
	defer func() { w.writerState = writerStateBody }()
	for k, v := range h {
		_, err := fmt.Fprintf(w.writer, "%s: %s\r\n", k, v)
		if err != nil {
			return err
		}
	}
	_, err := fmt.Fprint(w.writer, "\r\n")
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("cannot write body in state %d", w.writerState)
	}
	defer func() { w.writerState = writerStateTrailers }()
	return w.writer.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("cannot write chunked body in state %d", w.writerState)
	}
	totalWritten := 0

	n, err := fmt.Fprintf(w.writer, "%x\r\n", len(p))
	if err != nil {
		return totalWritten, err
	}
	totalWritten += n

	n, err = w.writer.Write(p)
	if err != nil {
		return totalWritten, err
	}
	totalWritten += n

	n, err = fmt.Fprint(w.writer, "\r\n")
	if err != nil {
		return totalWritten, err
	}
	totalWritten += n

	return totalWritten, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("cannot write chunked body done in state %d", w.writerState)
	}
	defer func() { w.writerState = writerStateTrailers }()
	return w.writer.Write([]byte("0\r\n"))
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.writerState != writerStateTrailers {
		return fmt.Errorf("cannot write headers in state %d", w.writerState)
	}
	defer func() { w.writerState = writerStateDone }()
	for k, v := range h {
		_, err := fmt.Fprintf(w.writer, "%s: %s\r\n", k, v)
		if err != nil {
			return err
		}
	}
	_, err := fmt.Fprint(w.writer, "\r\n")
	return err
}
