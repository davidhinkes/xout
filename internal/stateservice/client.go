package stateservice

import (
	"encoding/gob"
	"log"
	"net"
	"sync"
)

type Updater interface {
	Update(name string, value interface{})
}

// Client is a connection to the stateservice.
type Client struct {
	values       map[string]interface{}
	outgoingChan chan mutationMessage
	conn         net.Conn
	updater      Updater
	lock         sync.Mutex
}

// NewClient makes a new Client
func NewClient(conn net.Conn, updater Updater) *Client {
	client := &Client{
		values:       make(map[string]interface{}),
		outgoingChan: make(chan mutationMessage),
		conn:         conn,
		updater: updater,
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

// GetValues retrieves a particlar key's value from the stateservice.
// Nil will be returned in the event that the key does not exist.
func (c *Client) GetValues(names ...string) map[string]interface{} {
	ret := make(map[string]interface{})
	c.lock.Lock()
	defer c.lock.Unlock()
	for _, name := range names {
		v := c.values[name]
		if v == nil {
			continue
		}
		ret[name] = v
	}
	return ret
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
			log.Fatal(err)
		}
	}
}

func (c *Client) handleIncommingMessages() {
	decoder := gob.NewDecoder(c.conn)
	for {
		msg := mutationMessage{}
		if err := decoder.Decode(&msg); err != nil {
			log.Fatal(err)
		}
		c.lock.Lock()
		for name, value := range msg.Mutations {
			c.values[name] = value
		}
		c.lock.Unlock()
		if c.updater == nil {
			continue
		}
		for name, value := range msg.Mutations {
			c.updater.Update(name, value)
		}
	}
}
