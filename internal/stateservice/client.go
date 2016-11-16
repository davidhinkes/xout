package stateservice

import (
	"encoding/gob"
	"log"
	"net"
	"sync"
)

// Client is a connection to the stateservice.
type Client struct {
	values       map[string]interface{}
	outgoingChan chan mutationMessage
	conn         net.Conn
	lock         sync.Mutex
}

// NewClient makes a new Client
func NewClient(conn net.Conn) *Client {
	client := &Client{
		values:       make(map[string]interface{}),
		outgoingChan: make(chan mutationMessage),
		conn:         conn,
	}
	go client.handleOutgoingMessages()
	go client.handleIncommingMessages()
	return client
}

// GetValue retrieves a particlar key's value from the stateservice.
// Nil will be returned in the event that the key does not exist.
func (c *Client) GetValue(name string) interface{} {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.values[name]
}

// SetValues writes to the stateservices.
// Note that some time will need to pass before changes are reflected via GetValue.
func (c *Client) SetValues(mutations map[string]interface{}) {
	c.outgoingChan <- mutationMessage{
		Mutations: mutations,
	}
}

// SetValue is a convenience function for setting a single value. 
func (c *Client) SetValue(key string, value interface{}) {
	c.SetValues(map[string]interface{}{
		key: value,
	})
}

func (c *Client) handleOutgoingMessages() {
	encoder := gob.NewEncoder(c.conn)
	for msg := range c.outgoingChan {
		if err := encoder.Encode(msg); err != nil {
			log.Print(err)
		}
	}
}

func (c *Client) handleIncommingMessages() {
	decoder := gob.NewDecoder(c.conn)
	for {
		msg := mutationMessage{}
		if err := decoder.Decode(&msg); err != nil {
			log.Print(err)
		}
		c.lock.Lock()
		for name, value := range msg.Mutations {
			c.values[name] = value
		}
		c.lock.Unlock()
	}
}
