package services

import (
	"time"

	"github.com/gorilla/websocket"
)

// sendQueueSize is the per-connection outbound buffer. A full queue means the
// client is not draining fast enough; the message is then dropped (the full
// sync state self-heals on the next notification) instead of closing the
// connection. The socket is only closed when a write actually fails.
const sendQueueSize = 128

// writeTimeout bounds a single websocket write (and ping) so a stalled TCP
// peer cannot wedge a writer goroutine forever.
const writeTimeout = 10 * time.Second

// pingPeriod keeps connections alive through NATs/proxies that silently drop
// idle WebSockets (a very common cause of ~15s disconnects).
const pingPeriod = 10 * time.Second

// PongWait is how long the read side tolerates silence before declaring the
// peer dead. It must be comfortably larger than pingPeriod.
const PongWait = 30 * time.Second

// client wraps a websocket connection with a buffered outbound queue. A
// dedicated writer goroutine drains the queue and sends keepalive pings so
// broadcasts never block the game state mutex on a slow peer — a single stuck
// client can no longer freeze logins or actions for everyone else.
type client struct {
	conn *websocket.Conn
	send chan []byte
	done chan struct{}
}

func newClient(conn *websocket.Conn) *client {
	c := &client{
		conn: conn,
		send: make(chan []byte, sendQueueSize),
		done: make(chan struct{}),
	}
	go c.writeLoop()
	return c
}

// ConnectWS registers a raw upgraded websocket in the manager, wrapping it in
// a buffered client so broadcasts never block the game-loop mutex.
func (gm *GameManager) ConnectWS(pseudo string, conn *websocket.Conn, charInfo map[string]string) {
	gm.Connect(pseudo, newClient(conn), charInfo)
}

func (c *client) writeLoop() {
	defer c.conn.Close()
	pingTicker := time.NewTicker(pingPeriod)
	defer pingTicker.Stop()
	for {
		select {
		case <-c.done:
			return
		case msg := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-pingTicker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := c.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeTimeout)); err != nil {
				return
			}
		}
	}
}

// enqueue queues a payload for delivery without blocking. It reports whether
// the message was accepted; a full queue means the client is too slow and the
// message is dropped (the next full-state sync will bring it back up to date).
func (c *client) enqueue(msg []byte) bool {
	select {
	case c.send <- msg:
		return true
	default:
		return false
	}
}

// close stops the writer goroutine and the underlying socket.
func (c *client) close() {
	select {
	case <-c.done:
	default:
		close(c.done)
	}
	c.conn.Close()
}