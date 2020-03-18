package conn

import (
	"crypto/ed25519"
	"crypto/tls"
	"crypto/x509"

	"github.com/buuzcoin/go-buuzcoin/network"
	"github.com/lucas-clemente/quic-go"
	"github.com/pkg/errors"
)

// ErrInvalidCertificate is returned if server didn't send or sent invalid certificate
var ErrInvalidCertificate = errors.New("dial: Invalid certificate")

// Dial connects to UDP address provided.
// If connection succeeded, RemotePublicKey and RemoteAddress are set
func Dial(protoName, address string) (*Peer, error) {
	session, err := quic.DialAddr(address, &tls.Config{
		NextProtos:         []string{protoName},
		InsecureSkipVerify: true,
	}, nil)
	if err != nil {
		return nil, errors.Wrap(err, "quic.DialAddr failed")
	}

	if len(session.ConnectionState().PeerCertificates) != 1 {
		return nil, ErrInvalidCertificate
	}
	remoteCert := session.ConnectionState().PeerCertificates[0]
	if remoteCert.PublicKeyAlgorithm != x509.Ed25519 {
		return nil, ErrInvalidCertificate
	}
	remotePublicKey, ok := remoteCert.PublicKey.(ed25519.PublicKey)
	if !ok {
		return nil, ErrInvalidCertificate
	}

	peer := &Peer{
		RemotePublicKey: remotePublicKey,
		RemoteID:        network.DeriveAddress(remotePublicKey),

		session: session,
		done:    make(chan interface{}),
	}
	return peer, nil
}
