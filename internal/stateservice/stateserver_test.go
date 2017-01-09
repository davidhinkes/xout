package stateservice

import (
	"bytes"
	"encoding/gob"
	"net"
	"os"
	"testing"
	"time"
)

func encode(t *testing.T, e *gob.Encoder, v interface{}) {
	if err := e.Encode(v); err != nil {
		t.Fatal(err)
	}
}

func decode(t *testing.T, e *gob.Decoder, v interface{}) {
	if err := e.Decode(v); err != nil {
		t.Fatal(err)
	}
}

func TestGob(t *testing.T) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)
	encode(t, enc, mutationMessage{})
	decode(t, dec, &mutationMessage{})
	encode(t, enc, mutationMessage{
		Mutations: map[string]interface{}{
			"a": int(42),
		},
	})
	decode(t, dec, &mutationMessage{})
}

func TestGetSet(t *testing.T) {
	unixSocketPath := "/var/tmp/xout.unixsocket"
	listener, err := net.Listen("unix", unixSocketPath)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(unixSocketPath)
	go RunStateServer([]net.Listener{listener})
	conn, err := net.Dial("unix", unixSocketPath)
	if err != nil {
		t.Fatal(err)
	}
	client := NewClient(conn, nil)
	client.SetValues(map[string]interface{}{
		"abc": float64(42),
	})
	time.Sleep(1 * time.Millisecond)
	got, ok := client.GetValue("abc").(float64)
	if !ok {
		t.Error(ok)
	}
	if want := float64(42); got != want {
		t.Errorf("got %v != want %v", got, want)
	}
}

type simpleUpdater struct {
	nameLastUpdate  string
	valueLastUpdate interface{}
}

func (s *simpleUpdater) Update(n string, v interface{}) {
	s.nameLastUpdate = n
	s.valueLastUpdate = v
}

func TestUpdater(t *testing.T) {
	unixSocketPath := "/var/tmp/xout.unixsocket"
	listener, err := net.Listen("unix", unixSocketPath)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(unixSocketPath)
	go RunStateServer([]net.Listener{listener})
	conn, err := net.Dial("unix", unixSocketPath)
	if err != nil {
		t.Fatal(err)
	}
	updater := &simpleUpdater{}
	client := NewClient(conn, updater)
	client.SetValues(map[string]interface{}{
		"abc": float64(42),
	})
	// Ok, this is bad -- I know.  But I want to test the callback.
	time.Sleep(1 * time.Millisecond)
	if got, want := updater.nameLastUpdate, "abc"; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := updater.valueLastUpdate.(float64), float64(42); got != want {
		t.Errorf("got %v != want %v", got, want)
	}
}
