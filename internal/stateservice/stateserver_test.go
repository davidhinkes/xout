package stateservice

import (
	"bytes"
	"encoding/gob"
	"net"
	"os"
	"testing"
	"time"
)

func TestT(t *testing.T) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(mutationMessage{}); err != nil {
		t.Error(err)
	}
	err := enc.Encode(mutationMessage{
		Mutations: map[string]interface{}{
			"a": int(42),
		},
	})
	if err != nil {
		t.Error(err)
	}
}

func TestClientServer(t *testing.T) {
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
	client := NewClient(conn)
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
