package conn

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/lucas-clemente/quic-go"
	"github.com/pkg/errors"
)

func generateCertificate() (*tls.Certificate, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, errors.Wrap(err, "generateCertificate failed")
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"bzc-selfsigned"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, pub, priv)
	if err != nil {
		return nil, errors.Wrap(err, "generateCertificate failed")
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, errors.Wrap(err, "generateCertificate: failed to marshal private key")
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, errors.Wrap(err, "generateCertificate: creating certificate from PEM failed")
	}
	return &cert, nil
}

func TestNodeReadWrite(t *testing.T) {
	t.Run("QUIC listener", func(t *testing.T) {
		t.Parallel()
		cert, err := generateCertificate()
		if err != nil {
			t.Fatalf("Creating certificate failed: %+v\n", err)
		}

		listener, err := quic.ListenAddr("localhost:14050", &tls.Config{
			Certificates:       []tls.Certificate{*cert},
			NextProtos:         []string{"quic-transport-peer-test"},
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

		session, err := quic.DialAddr("localhost:14050", &tls.Config{
			NextProtos:         []string{"quic-transport-peer-test"},
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
