package net

import (
	"github.com/buuzcoin/go-buuzcoin/network/protocol"
)

// CurrentProtocolVersion is current version of protocol sent to remote
const CurrentProtocolVersion = 1

/*
	Initialization stage:
	1. Send HelloMessage to remote
	2. Recieve interlocutor's HelloMessage
	3. Send Ping message (for node with less ID)
	4. Send Pong message (node with greater ID)
	5. Add nodes to routing table
*/

// startPeerConnection perform initialization stage with remote peer
func (conn *Connection) startPeerConnection() {
	helloMessage := new(protocol.HelloMessage)
	helloMessage.ProtoVersion = 1
}
