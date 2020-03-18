package conn

import (
	"context"
	"crypto/ed25519"
	"encoding/binary"
	"errors"

	"github.com/golang/protobuf/proto"
	"github.com/lucas-clemente/quic-go"
)

// Peer represents connection with remote peer
type Peer struct {
	RemotePublicKey ed25519.PublicKey
	RemoteID        []byte

	session quic.Session
	stream  quic.Stream
	done    chan interface{}
}

// Close closes connection
func (peer *Peer) Close() {
	peer.session.Close()
	if peer.stream != nil {
		peer.stream.Close()
	}
	close(peer.done)
}

// Write sends raw data to remote peer.
// Returns false if error occured and connection was closed
func (peer *Peer) Write(data []byte) bool {
	if peer.stream == nil {
		var err error
		peer.stream, err = peer.session.OpenStreamSync(context.Background())
		if err != nil {
			peer.Close()
			return false
		}
	}

	buffer := make([]byte, 4, 4+len(data))
	binary.LittleEndian.PutUint32(buffer[:4], uint32(len(data)))
	buffer = append(buffer, data...)

	offset := 0
	for offset < len(buffer) {
		select {
		case <-peer.done:
			return false
		default:
			written, err := peer.stream.Write(buffer[offset:])
			if err != nil {
				peer.Close()
				return false
			}
			offset += written
		}
	}

	return true
}

func (peer *Peer) read(expectedLength int) []byte {
	if peer.stream == nil {
		var err error
		peer.stream, err = peer.session.AcceptStream(context.Background())
		if err != nil {
			peer.Close()
			return nil
		}
	}

	result := make([]byte, expectedLength)

	offset := 0
	for offset < expectedLength {
		select {
		case <-peer.done:
			return nil
		default:
			read, err := peer.stream.Read(result[offset:expectedLength])
			if err != nil {
				peer.Close()
				return nil
			}
			offset += read
		}
	}

	return result
}

// Read reads raw message from remote.
// nil is returned if connection was closed.
func (peer *Peer) Read() []byte {
	msgLenBuffer := peer.read(4)
	if msgLenBuffer == nil {
		return nil
	}

	messageLength := int(binary.LittleEndian.Uint32(msgLenBuffer))
	return peer.read(messageLength)
}

// ReadMessage reads raw message and message ID from remote.
// nil is returned if connection was closed.
func (peer *Peer) ReadMessage() (byte, []byte) {
	msgLenBuffer := peer.read(4)
	if msgLenBuffer == nil {
		return 0, nil
	}

	messageLength := int(binary.LittleEndian.Uint32(msgLenBuffer))
	rawMessage := peer.read(messageLength)
	return rawMessage[0], rawMessage[1:]
}

// ErrMessageEncodingFailed is returned by peer.Send function if message marshal has failed
var ErrMessageEncodingFailed = errors.New("peer: Protobuf message encoding failed")

// Send writes protobuf message with specific ID to remote peer
// If error returned is not ErrMessageEncodingFailed, connection was closed
func (peer *Peer) Send(messageID byte, message proto.Message) error {
	messageBytes, err := proto.Marshal(message)
	if err != nil {
		return ErrMessageEncodingFailed
	}
	success := peer.Write(append([]byte{messageID}, messageBytes...))
	if !success {
		return errors.New("peer.Send: write failed")
	}
	return nil
}
