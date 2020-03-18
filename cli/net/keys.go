package net

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"log"
	"math/big"
	"time"

	"github.com/bmatsuo/lmdb-go/lmdb"
	"github.com/buuzcoin/go-buuzcoin/network"
	"github.com/pkg/errors"
)

// SaveCertificate saves provided certificate and private key to local storage
func (netNode *NetworkNode) SaveCertificate(certPEM, keyPEM []byte) error {
	if err := netNode.localStorage.Env.Update(func(txn *lmdb.Txn) error {
		if err := txn.Put(netNode.localStorage.Node, []byte("TLSCert"), certPEM, 0); err != nil {
			return err
		}
		return txn.Put(netNode.localStorage.Node, []byte("TLSKey"), keyPEM, 0)
	}); err != nil {
		return errors.Wrap(err, "SaveCertificate: saving to local storage failed")
	}
	return nil
}

// GenerateCertificate creates new certificate and private key
func (netNode *NetworkNode) GenerateCertificate(pubKey, privKey []byte) (*tls.Certificate, error) {
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

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, pubKey, privKey)
	if err != nil {
		return nil, errors.Wrap(err, "GenerateCertificate failed")
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	privBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return nil, errors.Wrap(err, "GenerateCertificate: failed to marshal private key")
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, errors.Wrap(err, "LoadTLSConfig: parsing certificate's PEM blocks failed: %s")
	}
	return &tlsCert, nil
}

// LoadTLSConfig generates TLS configuration using node's ed25519 keys.
func (netNode *NetworkNode) LoadTLSConfig() error {
	tlsCert, err := netNode.GenerateCertificate(netNode.nodePubKey, netNode.nodePrivKey)
	if err != nil {
		return errors.Wrap(err, "LoadTLSConfig: GenerateCertificate failed")
	}

	netNode.tlsConfig = &tls.Config{
		Certificates:       []tls.Certificate{*tlsCert},
		NextProtos:         []string{"bzc-blockchain"},
		InsecureSkipVerify: true,
	}
	return nil
}

// LoadNodeKeys loads node's ECDSA public and private keys from local storage
// If they are not found, they are generated and saved
func (netNode *NetworkNode) LoadNodeKeys(forceRegenerate bool) error {
	var err error
	if !forceRegenerate {
		if err = netNode.localStorage.Env.View(func(txn *lmdb.Txn) error {
			var err error
			if netNode.nodePubKey, err = txn.Get(netNode.localStorage.Node, []byte("nodePubKey")); err != nil {
				return err
			}
			if netNode.nodePrivKey, err = txn.Get(netNode.localStorage.Node, []byte("nodePrivKey")); err != nil {
				return err
			}
			return nil
		}); err != nil && !lmdb.IsNotFound(err) {
			return errors.Wrap(err, "LoadNodeKeys: falied to load node's keys from local storage")
		}
	}

	if forceRegenerate || err != nil {
		// Generate and save new keys
		netNode.nodePubKey, netNode.nodePrivKey, err = ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return errors.Wrap(err, "LoadNodeKeys: falied to generate new key")
		}
		if err = netNode.localStorage.Env.Update(func(txn *lmdb.Txn) error {
			if err := txn.Put(netNode.localStorage.Node, []byte("nodePubKey"), netNode.nodePubKey, 0); err != nil {
				return err
			}
			if err := txn.Put(netNode.localStorage.Node, []byte("nodePrivKey"), netNode.nodePrivKey, 0); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return errors.Wrap(err, "LoadNodeKeys: falied to save node keys to local storage")
		}
	}

	netNode.nodeAddress = network.DeriveAddress(netNode.nodePubKey)
	log.Printf("Node ID: 0x%s\n", hex.EncodeToString(netNode.nodeAddress))
	return nil
}
