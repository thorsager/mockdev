package rawhttp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
)

type Writer struct {
	wtr  http.ResponseWriter
	done bool
}

func (w *Writer) Write(raw []byte) (int, error) {
	if w.done {
		return -1, fmt.Errorf("already used, can only write one response pr. writer")
	}

	rdr := bufio.NewReader(bytes.NewReader(raw))
	resp, err := http.ReadResponse(rdr, nil)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close() // error is ignored
	w.done = true
	for h, vs := range resp.Header {
		for _, v := range vs {
			w.wtr.Header().Add(h, v)
		}
	}

	w.wtr.WriteHeader(resp.StatusCode)
	written, err := io.Copy(w.wtr, resp.Body)
	return int(written), err
}

func NewWriter(w http.ResponseWriter) io.Writer {
	return &Writer{wtr: w}
}
