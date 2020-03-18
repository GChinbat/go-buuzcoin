package conn

import (
	"context"
	"crypto/tls"
	"testing"

	"github.com/lucas-clemente/quic-go"
)

func TestConnect(t *testing.T) {
	t.Run("QUIC listener", func(t *testing.T) {
		t.Parallel()
		cert, err := generateCertificate()
		if err != nil {
			t.Fatalf("Creating certificate failed: %+v\n", err)
		}

		listener, err := quic.ListenAddr("localhost:14052", &tls.Config{
			Certificates:       []tls.Certificate{*cert},
			NextProtos:         []string{"quic-transport-connect-test"},
			InsecureSkipVerify: true,
		}, nil)
		if err != nil {
			t.Fatalf("quic.ListenAddr failed: %+v\n", err)
		}
		defer listener.Close()

		session, err := listener.Accept(context.Background())
		if err != nil {
			t.Fatalf("listener.Accept failed: %+v\n", err)
		}

		peer := &Peer{
			session: session,
			done:    make(chan interface{}),
		}
		defer peer.Close()

		if message := peer.Read(); message == nil || string(message) != "Hello!" {
			t.Fatal("peer.Read failed\n")
		}

		if !peer.Write([]byte("test")) {
			t.Fatal("peer.Write failed\n")
		}
	})

	t.Run("QUIC client", func(t *testing.T) {
		t.Parallel()

		peer, err := Dial("quic-transport-connect-test", "localhost:14052")
		if err != nil {
			t.Fatalf("Dial failed: %+v", err)
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
