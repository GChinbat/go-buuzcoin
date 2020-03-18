package conn

import (
	"crypto/tls"
	"net"

	"github.com/lucas-clemente/quic-go"
	"github.com/pixelbender/go-stun/stun"
	"github.com/pkg/errors"
)

// QUICTransport maintains incoming and outgoing connections
type QUICTransport struct {
	Address             net.Addr
	ProtocolName        string
	TLSCertificate      tls.Certificate
	IncomingConnections chan *Peer

	packetConn net.PacketConn
	done       chan interface{}
}

// Close closes all channels and terminates QUIC listener
func (transport *QUICTransport) Close() {
	close(transport.IncomingConnections)
	// This implicitcly closes QUIC listener
	close(transport.done)
}

// Init discovers own IP address using STUN protocol and creates QUIC listener
func (transport *QUICTransport) Init(stunServer string) error {
	transport.done = make(chan interface{})

	if len(stunServer) == 0 {
		stunServer = "stun.l.google.com:19302"
	}

	var err error
	transport.packetConn, transport.Address, err = stun.Discover("stun:" + stunServer)
	if err != nil {
		return errors.Wrap(err, "NetworkNode.Init: STUN discovery failed")
	}

	listener, err := quic.Listen(transport.packetConn, &tls.Config{
		Certificates:       []tls.Certificate{transport.TLSCertificate},
		NextProtos:         []string{transport.ProtocolName},
		InsecureSkipVerify: true,
	}, nil)
	if err != nil {
		return errors.Wrap(err, "NetworkNode.Init: creating QUIC listener failed")
	}

	go transport.listen(listener)
	return nil
}
