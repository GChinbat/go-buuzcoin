package protocol

const (
	// MessageGetBlockHeaders is ID for GetBlockHeaders message
	MessageGetBlockHeaders byte = 0x01
	// MessageBlockHeaders is ID for BlockHeaders message
	MessageBlockHeaders byte = 0x02
	// MessageNewBlockHashes is ID for NewBlockHashes message
	MessageNewBlockHashes byte = 0x03
	// MessageGetTransactions is ID for GetTransactions message
	MessageGetTransactions byte = 0x04
	// MessageTransactions is ID for Transactions message
	MessageTransactions byte = 0x05
	// MessageNewBlock is ID for NewBlock message
	MessageNewBlock byte = 0x06
	// MessageGetNodeData is ID for GetNodeData message
	MessageGetNodeData byte = 0x07
	// MessageNodeData is ID for NodeData message
	MessageNodeData byte = 0x08

	// MessagePing is ID for Ping message
	MessagePing byte = 0xF0
	// MessagePong is ID for Pong message
	MessagePong byte = 0xF1
	// MessageFindNode is ID for FindNode message
	MessageFindNode byte = 0xF2
	// MessageNeighbours is ID for Neighbours message
	MessageNeighbours byte = 0xF3
)
