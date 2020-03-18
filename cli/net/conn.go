package net

import (
	"context"
	"encoding/binary"

	"github.com/lucas-clemente/quic-go"
)

// Connection represents network connection
type Connection struct {
	done chan interface{}

	session quic.Session
	netNode *NetworkNode

	outgoingStream quic.SendStream
	incomingStream quic.ReceiveStream

	incomingRawMessages chan []byte
	outgoingRawMessages chan []byte

	RemoteID        []byte
	RemotePublicKey []byte
}

func (conn *Connection) listenLoop() {
	n := 0
	buffer := make([]byte, 4096)

	var writeChan chan []byte
	for {
		if n > 0 {
			writeChan = conn.incomingRawMessages
		} else {
			writeChan = nil
		}
		select {
		case <-conn.done:
			return
		case writeChan <- buffer[:n]:
			n = 0
			continue
		default:
			if n == 0 {
				var err error
				n, err = conn.incomingStream.Read(buffer)
				if err != nil {
					close(conn.done)
					return
				}
			}
		}
	}
}
func (conn *Connection) writeLoop() {
	var outgoingRawMessages chan []byte
	for {
		select {
		case outgoingRawMessage := <-outgoingRawMessages:
			_, err := conn.outgoingStream.Write(outgoingRawMessage)
			if err != nil {
				close(conn.done)
				return
			}
		case <-conn.done:
			return
		}
	}
}

// Read blocks goroutine and reads data from connection
func (conn *Connection) Read() []byte {
	offset := 0
	lenBuffer := make([]byte, 4)

	var data []byte
	for offset < 4 {
		select {
		case data = <-conn.incomingRawMessages:
			copy(lenBuffer[offset:], data)
			offset += len(data)
		case <-conn.done:
			return nil
		}
	}
	offset -= 4

	messageLength := int(binary.LittleEndian.Uint32(lenBuffer))
	buffer := make([]byte, messageLength)
	for offset < messageLength {
		select {
		case data = <-conn.incomingRawMessages:
			copy(buffer[offset:], data)
			offset += len(data)
		case <-conn.done:
			return nil
		}
	}

	return buffer
}

// Write blocks goroutine and writes data to peer
func (conn *Connection) Write(data []byte) bool {
	lenBuffer := make([]byte, 4)
	binary.LittleEndian.PutUint32(lenBuffer, uint32(len(data)))

loop:
	for {
		select {
		case conn.outgoingRawMessages <- lenBuffer:
			break loop
		case <-conn.done:
			return false
		}
	}

	offset := 0
	for offset < len(data) {
		var remainingLength int
		if len(data)-offset < 1024 {
			remainingLength = len(data) - offset
		} else {
			remainingLength = 1024
		}
		select {
		case conn.outgoingRawMessages <- data[offset : offset+remainingLength]:
			offset += remainingLength
		case <-conn.done:
			return false
		}
	}

	return true
}

// HandleConnection handles new connection and assignes state to it
func (netNode *NetworkNode) HandleConnection(sess quic.Session) {
	defer sess.Close()
	conn := &Connection{
		done:    make(chan interface{}),
		session: sess,
		netNode: netNode,
	}

	outgoingStream, err := sess.OpenUniStream()
	if err != nil {
		return
	}
	defer outgoingStream.Close()
	conn.outgoingStream = outgoingStream

	incomingStream, err := sess.AcceptUniStream(context.Background())
	if err != nil {
		return
	}
	conn.incomingStream = incomingStream

	go conn.writeLoop()
	go conn.listenLoop()

	for {
		select {
		case <-netNode.done:
			close(conn.done)
			return
		}
	}
}
