package conn

import (
	"crypto/tls"
	"testing"

	"github.com/lucas-clemente/quic-go"
)

func TestQUICTransport(t *testing.T) {
	cert, err := generateCertificate()
	if err != nil {
		t.Fatalf("generateCertificate failed: %+v\n", err)
	}

	transport := &QUICTransport{
		ProtocolName:        "QUICTransport-test",
		TLSCertificate:      *cert,
		IncomingConnections: make(chan *Peer),
		done:                make(chan interface{}),
	}
	listener, err := quic.ListenAddr("localhost:14051", &tls.Config{
		Certificates:       []tls.Certificate{*cert},
		NextProtos:         []string{"quic-transport-test"},
		InsecureSkipVerify: true,
	}, nil)
	if err != nil {
		t.Fatalf("quic.ListenAddr failed: %+v\n", err)
	}

	t.Run("QUICTransport server", func(t *testing.T) {
		t.Parallel()
		go transport.listen(listener)
		defer transport.Close()

		peer := <-transport.IncomingConnections
		defer peer.Close()

		if message := peer.Read(); message == nil || string(message) != "Hello!" {
			t.Fatal("peer.Read failed")
		}

		if !peer.Write([]byte("test")) {
			t.Fatal("peer.Write failed")
		}
	})
	t.Run("QUIC client", func(t *testing.T) {
		t.Parallel()

		session, err := quic.DialAddr("localhost:14051", &tls.Config{
			NextProtos:         []string{"quic-transport-test"},
			InsecureSkipVerify: true,
		}, nil)
		if err != nil {
			t.Fatalf("quic.DialAddr failed: %+v\n", err)
		}

		peer := &Peer{
			session: session,
			done:    make(chan interface{}),
		}
		defer peer.Close()

		if !peer.Write([]byte("Hello!")) {
			t.Fatal("peer.Write failed\n")
		}

		if message := peer.Read(); message == nil || string(message) != "test" {
			t.Fatal("peer.Read failed\n")
		}
	})
}
