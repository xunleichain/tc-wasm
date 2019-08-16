package types

import (
	"math/big"
)

// StateDB is an VM database for full state querying.
type StateDB interface {
	CreateAccount(Address)

	SubBalance(Address, *big.Int)
	AddBalance(Address, *big.Int)
	GetBalance(Address) *big.Int

	GetNonce(Address) uint64
	SetNonce(Address, uint64)

	GetCodeHash(Address) Hash
	GetContractCode([]byte) []byte
	GetCode(Address) []byte
	SetCode(Address, []byte)
	GetCodeSize(Address) int
	IsContract(Address) bool

	AddRefund(uint64)
	GetRefund() uint64

	GetState(Address, Hash) []byte
	SetState(Address, Hash, []byte)

	Suicide(Address) bool
	HasSuicided(Address) bool

	// Exist reports whether the given account exists in state.
	// Notably this should also return true for suicided accounts.
	Exist(Address) bool
	// Empty returns whether the given account is empty. Empty
	// is defined according to EIP161 (balance = nonce = code = 0).
	Empty(Address) bool

	RevertToSnapshot(int)
	Snapshot() int

	AddLog(*Log)
	AddPreimage(Hash, []byte)

	ForEachStorage(Address, func(Hash, []byte) bool)
	TxHash() Hash
	Logs() []*Log

	SubTokenBalance(addr Address, token Address, amount *big.Int)
	AddTokenBalance(addr Address, token Address, amount *big.Int)
	GetTokenBalance(addr Address, token Address) *big.Int
	GetTokenBalances(addr Address) TokenValues
}

type TokenValue struct {
	TokenAddr Address  `json:"tokenAddress"`
	Value     *big.Int `json:"value"`
}

type TokenValues []TokenValue

type ChainContext interface {
	// GetHeader returns the hash corresponding to their hash.
	// GetHeader(uint64) *Header
}

type ContractRef interface {
	// Address() Address
}

// type OpCode byte

type Contract interface {
	// AsDelegate() Contract
	// GetOp(n uint64) OpCode
	// GetByte(n uint64) byte
	// Caller() Address
	// UseGas(gas uint64) (ok bool)
	// Address() Address
	// Value() *big.Int
	// SetCode(hash Hash, code []byte)
	// SetCallCode(addr *Address, hash Hash, code []byte)

	// //
	// GetCode() []byte
	// GetCodeHash() Hash
	// GetCodeAddr() *Address
	// GetInput() []byte
	// SetInput(input []byte)
	// GetGas() uint64
}

type Interpreter interface {
	// Run(contract Contract, input []byte) (ret []byte, err error)
}

type VmConfig interface {
}

// type (
// 	// CanTransferFunc is the signature of a transfer guard function
// 	CanTransferFunc func(StateDB, Address, *big.Int) bool
// 	// TransferFunc is the signature of a transfer function
// 	TransferFunc func(StateDB, Address, Address, *big.Int)
// 	// GetHashFunc returns the nth block hash in the blockchain
// 	// and is used by the BLOCKHASH EVM op code.
// 	GetHashFunc func(uint64) Hash
// )

// Context provides the EVM with auxiliary information. Once provided
// it shouldn't be modified.
type Context interface {
	// // CanTransferFunc is the signature of a transfer guard function
	// CanTransferFunc(StateDB, Address, *big.Int) bool
	// // TransferFunc is the signature of a transfer function
	// TransferFunc(StateDB, Address, Address, *big.Int)
	// // GetHashFunc returns the nth block hash in the blockchain
	// // and is used by the BLOCKHASH EVM op code.
	// GetHashFunc(uint64) Hash

	// // Message information
	// // Origin   Address // Provides information for ORIGIN
	// // GasPrice *big.Int       // Provides information for GASPRICE

	// // // Block information
	// // Coinbase    Address // Provides information for COINBASE
	// // GasLimit    uint64         // Provides information for GASLIMIT
	// // BlockNumber *big.Int       // Provides information for NUMBER
	// // Time        *big.Int       // Provides information for TIME
	// // Difficulty  *big.Int       // Provides information for DIFFICULTY
	// GetOrigin() Address
	// GetGasPrice() *big.Int
	// GetCoinbase() Address
	// GetGasLimit() uint64
	// GetBlockNumber() *big.Int
	// GetTime() *big.Int
	// GetDifficulty() *big.Int
}
