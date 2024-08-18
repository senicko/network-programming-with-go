package ch04

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	BinaryType uint8 = iota + 1
	StringType
	MaxPayloadSize uint32 = 10 << 20 // 10 MB
)

var ErrMaxPayloadSizeExceeded = errors.New("maximum payload size exceeded")

type Payload interface {
	fmt.Stringer
	io.ReaderFrom
	io.WriterTo
	Bytes() []byte
}

// Binary type

type Binary []byte

func (m Binary) Bytes() []byte  { return m }
func (m Binary) String() string { return string(m) }

func (m Binary) WriteTo(w io.Writer) (int64, error) {
	err := binary.Write(w, binary.BigEndian, BinaryType) // header
	if err != nil {
		return 0, err
	}
	var n int64 = 1

	err = binary.Write(w, binary.BigEndian, uint32(len(m))) // length
	if err != nil {
		return n, err
	}
	n += 4

	o, err := w.Write(m) // payload
	return n + int64(o), err
}

func (m *Binary) ReadFrom(r io.Reader) (int64, error) {
	var size uint32
	err := binary.Read(r, binary.BigEndian, &size)
	if err != nil {
		return 0, err
	}
	var n int64 = 4
	if size > MaxPayloadSize {
		return n, ErrMaxPayloadSizeExceeded
	}

	*m = make([]byte, size)
	o, err := r.Read(*m)

	return n + int64(o), err
}

// String type

type String string

func (m String) Bytes() []byte  { return []byte(m) }
func (m String) String() string { return string(m) }

func (m String) WriteTo(w io.Writer) (int64, error) {
	err := binary.Write(w, binary.BigEndian, StringType) // header
	if err != nil {
		return 0, err
	}
	var n int64 = 1

	err = binary.Write(w, binary.BigEndian, uint32(len(m))) // length
	if err != nil {
		return n, err
	}
	n += 4

	o, err := w.Write([]byte(m)) // payload
	return n + int64(o), err
}

func (m *String) ReadFrom(r io.Reader) (int64, error) {
	var size uint32
	err := binary.Read(r, binary.BigEndian, &size)
	if err != nil {
		return 0, err
	}
	var n int64 = 4
	if size > MaxPayloadSize {
		return n, ErrMaxPayloadSizeExceeded
	}

	buf := make([]byte, size)
	o, err := r.Read(buf)
	if err != nil {
		return n, err
	}

	*m = String(buf)
	return n + int64(o), nil
}

func decode(r io.Reader) (Payload, error) {
	var t uint8
	err := binary.Read(r, binary.BigEndian, &t)
	if err != nil {
		return nil, err
	}

	var payload Payload

	switch t {
	case BinaryType:
		payload = new(Binary)
	case StringType:
		payload = new(String)
	default:
		return nil, errors.New("unknown data type")
	}

	// At this point we already know what type are we reading, so the ReadFrom itself does not
	// need this information.
	_, err = payload.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	return payload, nil
}
