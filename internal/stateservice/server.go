package stateservice

import (
	"encoding/gob"
	"log"
	"net"
)

type state struct {
	clients   []*client
	entries   map[string]entry
	timestamp uint64
}

type client struct {
	c          chan mutationMessage
	canceled   bool
	lastUpdate uint64
}

type entry struct {
	lastUpdate uint64
	value      interface{}
}

// RunStateServer starts a stateservice server and listens to the specified network
// interfaces.
// This function does not return.
func RunStateServer(listeners []net.Listener) error {
	s := &state{
		entries: make(map[string]entry),
	}
	newConnectionsChan := make(chan net.Conn, 256)
	incommingRequestsChan := make(chan mutationMessage, 256)
	cancelChan := make(chan *client, 16)
	for _, listener := range listeners {
		go handleListener(newConnectionsChan, listener)
	}
	for {
		select {
		case newConn := <-newConnectionsChan:
			go handleIncommingClientMessages(incommingRequestsChan, newConn)
			c := make(chan mutationMessage, 256)
			client := &client{c: c}
			go handleOutgoingClientMessages(newConn, client, cancelChan)
			s.clients = append(s.clients, client)
		case req := <-incommingRequestsChan:
			s.timestamp++
			for name, value := range req.Mutations {
				s.entries[name] = entry{
					lastUpdate: s.timestamp,
					value:      value,
				}
			}
			s.updateClients()
		case client := <-cancelChan:
			if !client.canceled {
				client.canceled = true
				close(client.c)
			}
		}
	}
	return nil
}

func (s *state) updateClients() {
	for _, client := range s.clients {
		if client.canceled {
			continue
		}
		msg := mutationMessage{
			Mutations: make(map[string]interface{}),
		}
		for name, entry := range s.entries {
			if entry.lastUpdate > client.lastUpdate {
				msg.Mutations[name] = entry.value
			}
		}
		client.lastUpdate = s.timestamp
		if len(msg.Mutations) == 0 {
			continue
		}
		client.c <- msg
	}
}

func handleOutgoingClientMessages(conn net.Conn, client *client, cancelRequest chan<- *client) {
	enc := gob.NewEncoder(conn)
	for msg := range client.c {
		if err := enc.Encode(msg); err != nil {
			log.Print(err)
			conn.Close()
			cancelRequest <- client
			break
		}
	}
	conn.Close()
}

func handleIncommingClientMessages(c chan<- mutationMessage, conn net.Conn) {
	dec := gob.NewDecoder(conn)
	for {
		msg := mutationMessage{}
		if err := dec.Decode(&msg); err != nil {
			log.Print(err)
			break
		}
		c <- msg
	}
	conn.Close()
}

func handleListener(c chan<- net.Conn, listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			listener.Close()
			return
		}
		c <- conn
	}
}
