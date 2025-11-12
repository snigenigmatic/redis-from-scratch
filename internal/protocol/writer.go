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

// TODO: Writer covers the main RESP types. If you add complex types (e.g., nested arrays
// or custom serialization for zset members), add helper methods and tests here.

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

// WriteZsetMember writes a zset member as a two-element array: [score, member]
// Format: *2\r\n$N\r\nscore\r\n$M\r\nmember\r\n
func (w *Writer) WriteZsetMember(score float64, member string) error {
	if _, err := fmt.Fprintf(w.w, "*2\r\n"); err != nil {
		return err
	}
	if err := w.WriteBulkString(fmt.Sprintf("%f", score)); err != nil {
		return err
	}
	if err := w.WriteBulkString(member); err != nil {
		return err
	}
	return nil
}

// WriteNestedArray writes an array containing a cursor string and a sub-array of keys
// Format: *2\r\n$N\r\ncursor\r\n*M\r\n...keys...
func (w *Writer) WriteNestedArray(cursor string, keys []string) error {
	// Write outer array of 2 elements
	if _, err := fmt.Fprintf(w.w, "*2\r\n"); err != nil {
		return err
	}
	// Write cursor as bulk string
	if err := w.WriteBulkString(cursor); err != nil {
		return err
	}
	// Write keys as inner array
	if _, err := fmt.Fprintf(w.w, "*%d\r\n", len(keys)); err != nil {
		return err
	}
	for _, key := range keys {
		if err := w.WriteBulkString(key); err != nil {
			return err
		}
	}
	return nil
}
