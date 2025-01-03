package client

import (
	"log/slog"
	"net"
)

type ClientMsg struct {
	Conn net.Conn
	Msg  []byte
}

type Client struct {
	conn      net.Conn
	msgChan   chan ClientMsg
	closeChan chan net.Conn
}

func NewClient(conn net.Conn, msgChan chan ClientMsg, closeChan chan net.Conn) *Client {
	return &Client{
		conn:      conn,
		msgChan:   msgChan,
		closeChan: closeChan,
	}
}

/*
* Responsble for reading oncomming messages from client and sending them back to the server
 */
func (c *Client) ReadLoop() {
	for {
		buff := make([]byte, 1024)
		n, err := c.conn.Read(buff)
		if err != nil {
			slog.Error("error reading from connection", "err", err)
			// TODO: Send message back to server to close connection
			return
		}

		command := buff[:n]
		// Send command back to server
		c.msgChan <- ClientMsg{
			msg:  command,
			conn: c.conn,
		}
	}
}
