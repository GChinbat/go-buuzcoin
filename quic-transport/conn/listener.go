package conn

import (
	"context"

	"github.com/lucas-clemente/quic-go"
)

func (transport *QUICTransport) listen(listener quic.Listener) {
	defer listener.Close()

	var (
		peer            *Peer
		incomingSession chan *Peer
	)
	for {
		if peer == nil {
			incomingSession = nil
		} else {
			incomingSession = transport.IncomingConnections
		}

		select {
		case <-transport.done:
			return
		case incomingSession <- peer:
			peer = nil
			continue
		default:
			if peer != nil {
				continue
			}

			session, err := listener.Accept(context.Background())
			if err != nil {
				continue
			}
			peer = &Peer{
				session: session,
				done:    make(chan interface{}),
			}
		}
	}
}
