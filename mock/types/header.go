package types

// Header represents a block header.
type Header struct {
	ParentHash Hash
	Coinbase   Address
	GasLimit   uint64
	Time       uint64
	Height     uint64
}
