package protocol

import (
	"fmt"
	"io"
)

type Writer struct {
	w io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w}
}

func (w *Writer) WriteSimpleString(s string) error {
	_, err := fmt.Fprintf(w.w, "+%s\r\n", s)
	return err
}

func (w *Writer) WriteError(s string) error {
	_, err := fmt.Fprintf(w.w, "-%s\r\n", s)
	return err
}

func (w *Writer) WriteInteger(n int) error {
	_, err := fmt.Fprintf(w.w, ":%d\r\n", n)
	return err
}

func (w *Writer) WriteBulkString(s string) error {
	_, err := fmt.Fprintf(w.w, "$%d\r\n%s\r\n", len(s), s)
	return err
}

func (w *Writer) WriteNull() error {
	_, err := fmt.Fprintf(w.w, "$-1\r\n")
	return err
}

func (w *Writer) WriteArray(arr []string) error {
	if _, err := fmt.Fprintf(w.w, "*%d\r\n", len(arr)); err != nil {
		return err
	}
	for _, s := range arr {
		if err := w.WriteBulkString(s); err != nil {
			return err
		}
	}
	return nil
}
