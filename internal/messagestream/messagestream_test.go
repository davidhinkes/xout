package messagestream

import (
	"bytes"
	"testing"
)

func stringToCodec(str string) *bytesCodec {
	return &bytesCodec{
		bytes: []byte(string(str)),
	}
}

func TestX(t *testing.T) {
	buf := &bytes.Buffer{}
	writer := Writer{
		Writer: buf,
	}
	writer.Write(stringToCodec("abc"))
	if want, got := 4+16+3, buf.Len(); got != want {
		t.Errorf("need the number of bytes to be %v, not %v", want, got)
	}
	writer.Write(stringToCodec("de"))
	reader := Reader{
		Reader: buf,
	}
	var next bytesCodec
	if err := reader.Next(&next); err != nil {
		t.Error(err)
	}
	if got, want := string(next.bytes), "abc"; got != want {
		t.Errorf("got %v, got %v", want, got)
	}
	if err := reader.Next(&next); err != nil {
		t.Error(err)
	}
	if got, want := string(next.bytes), "de"; got != want {
		t.Errorf("got %v, got %v", want, got)
	}
}

type bytesCodec struct {
	bytes []byte
}

func (b bytesCodec) Encode() []byte {
	return b.bytes
}

func (b *bytesCodec) Decode(bytes []byte) error {
	b.bytes = bytes
	return nil
}
