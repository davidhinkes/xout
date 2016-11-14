// Package messagestream implements the message stream protocol for exchanging
// messages over io.Readers and io.Writers.
// TODO: describe the wire-protocol.
package messagestream

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"io"
)

const (
	headerLength = 4 + int(md5.Size)
)

var (
	endian = binary.LittleEndian
)

// Codec represents a type that can be encoded and decoded.
type Codec interface {
	Decode([]byte) error
	Encode() []byte
}

// Writer is an object that can send a message over an io.Writer.
type Writer struct {
	Writer io.Writer
}

// Write encodes codec and sends over Writer.Writer.
func (w Writer) Write(codec Codec) error {
	header := make([]byte, headerLength)
	data := codec.Encode()
	md5Sum := md5.Sum(data)
	len := uint32(len(data))
	endian.PutUint32(header, len)
	copy(header[4:], md5Sum[:])
	if _, err := w.Writer.Write(header); err != nil {
		return err
	}
	if _, err := w.Writer.Write(data); err != nil {
		return err
	}
	return nil
}

// Reader is an object that can materalize messages from an io.Reader.
type Reader struct {
	Reader io.Reader
	buffer bytes.Buffer
}

func (r *Reader) fill(codec Codec) (bool, error) {
	_, err := r.buffer.ReadFrom(r.Reader)
	if err != nil {
		return false, err
	}
	if r.buffer.Len() < headerLength {
		return false, nil
	}
	dataSize := int(endian.Uint32(r.buffer.Bytes()))
	md5Sum := r.buffer.Bytes()[4:headerLength]
	if r.buffer.Len() < headerLength+dataSize {
		return false, nil
	}
	defer r.buffer.Next(headerLength + dataSize)
	data := r.buffer.Bytes()[headerLength:(headerLength + dataSize)]
	if s := md5.Sum(data); bytes.Compare(s[:], md5Sum) != 0 {
		return false, nil
	}
	if err := codec.Decode(data); err != nil {
		return false, nil
	}
	return true, nil
}

// Next fills codec and will block for data to arrive.
// Errors will be returned for io.Reader errors only.
// Codec decode errors will be ignored.
func (r *Reader) Next(codec Codec) error {
	for {
		ok, err := r.fill(codec)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
	}
}
