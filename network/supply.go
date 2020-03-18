package network

// This file describes coin supply rules: per-block reward and unit naming convention

const (
	// Wei is one unit of Buuzcoin blockchain
	Wei uint64 = 1
	// Gwei is unit of Buuzcoin blockchain
	Gwei uint64 = 10000
	// Buuz is unit of Buuzcoin blockchain
	Buuz uint64 = 100000000
)

const (
	// EraBlocks shows after how much blocks reward will be halved
	EraBlocks uint64 = 210000
	// InitialReward is reward for blocks of first era
	InitialReward uint64 = 50 * Buuz
)

// BlockReward calculates reward for block with specific index
func BlockReward(index uint64) uint64 {
	var era = index / EraBlocks
	// InitialReward / 2 ** (era - 1)
	return InitialReward >> era
}
