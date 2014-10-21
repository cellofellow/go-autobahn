package autobahn

import (
	"errors"
	"github.com/gopherjs/gopherjs/js"
)

var autobahn js.Object = js.Global.Get("autobahn")

type Connection struct {
	js.Object
	closeChannels []chan<- CloseMessage
}

type CloseMessage struct {
	Reason       string
	CloseReason  string
	CloseMessage string
}

// Optional options pair.
type Option struct {
	key   string
	value js.Object
}

// Create a new autobahn.Connection object.
// Options documentation http://autobahn.ws/js/reference.html#connection-options
func NewConnection(url, realm string, options ...Option) *Connection {
	callopts := map[string]interface{}{"url": url, "realm": realm}
	for _, o := range options {
		callopts[o.key] = o.value
	}
	jsconn := autobahn.Get("Connection").New(callopts)
	conn := Connection{jsconn, make([]chan<- CloseMessage, 0)}
	conn.Set("onclose", conn.onClose)
	return &conn
}

// Open a connection. Creates a Cession.
func (c *Connection) Open() *Session {
	ch := make(chan *Session)
	c.Set("onopen", func(session js.Object) {
		ch <- &Session{session}
	})
	c.Call("open")
	ret := <-ch
	c.Delete("onopen")
	return ret
}

// Close a connection.
// Optional options: realm and message.
func (c *Connection) Close(options ...Option) error {
	var reason, message string
	for _, o := range options {
		switch o.key {
		case "reason":
			reason = o.value.Str()
		case "message":
			message = o.value.Str()
		}
	}
	ret := c.Call("close", map[string]string{"reason": reason, "message": message})
	if !ret.IsUndefined() {
		return errors.New(ret.Str())
	}
	return nil
}

func (c *Connection) AddCloseListener(ch chan<- CloseMessage) {
	c.closeChannels = append(c.closeChannels, ch)
}

func (c *Connection) onClose(reason string, details js.Object) {
	var message CloseMessage
	if reason == "close" {
		message = CloseMessage{
			reason,
			details.Get("reason").Str(),
			details.Get("message").Str(),
		}
	} else {
		message = CloseMessage{reason, "", ""}
	}
	for _, ch := range c.closeChannels {
		ch <- message
	}
}

func (c *Connection) GetSession() *Session {
	return &Session{c.Get("session")}
}

func (c *Connection) IsConnected() bool {
	return c.Get("isConnected").Bool()
}

func (c *Connection) IsOpen() bool {
	return c.Get("isOpen").Bool()
}

func (c *Connection) IsRetrying() bool {
	return c.Get("isRetrying").Bool()
}

func (c *Connection) GetTransport() *Transport {
	return &Transport{c.Get("transport")}
}
