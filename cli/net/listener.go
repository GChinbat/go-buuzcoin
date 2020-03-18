package net

import (
	"context"

	quic "github.com/lucas-clemente/quic-go"
	"github.com/pkg/errors"
)

// Listen opens node listener on specified address
func (netNode *NetworkNode) Listen(addr string) error {
	listener, err := quic.ListenAddr(addr, netNode.tlsConfig, nil)
	if err != nil {
		return errors.Wrap(err, "Listen: starting QUIC listener failed")
	}
	defer listener.Close()

	incomingSession := make(chan quic.Session)
	defer close(incomingSession)

	go func() {
		var (
			conn quic.Session = nil
			err  error
		)
		var newSession chan quic.Session
		for {
			if err == nil && conn != nil {
				newSession = incomingSession
			} else {
				newSession = nil
			}
			select {
			case <-netNode.done:
				return
			case newSession <- conn:
				continue
			default:
				conn, err = listener.Accept(context.Background())
			}
		}
	}()

	for {
		select {
		case session := <-incomingSession:
			go netNode.HandleConnection(session)
		case <-netNode.done:
			return nil
		}
	}
}
