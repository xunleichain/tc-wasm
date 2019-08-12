package vm

import "errors"

var (
	ErrMemoryGet          = errors.New("memory get* failed")
	ErrMemorySet          = errors.New("memory set* failed")
	ErrGasOverflow        = errors.New("gas overflow (uint64)")
	ErrContractAbort      = errors.New("contract abort")
	ErrInvalidApiArgs     = errors.New("invalid api args")
	ErrBalanceNotEnough   = errors.New("insufficient balance in api")
	ErrContractNotPayable = errors.New("contract not payable")
	ErrInvalidEnvArgs     = errors.New("invalid env args")
	ErrMallocMemory       = errors.New("malloc() failed in api")

	ErrCallDepth                = errors.New("vm: max call depth exceeded")
	ErrContractNoCode           = errors.New("vm: contract no code")
	ErrCodeStoreOutOfGas        = errors.New("vm: contract creation code storage out of gas")
	ErrExecutionReverted        = errors.New("vm: execution reverted")
	ErrMaxCodeSizeExceeded      = errors.New("vm: max code size exceeded")
	ErrInsufficientBalance      = errors.New("vm: insufficient balance in transfer")
	ErrContractAddressCollision = errors.New("vm: contract address collision")
	ErrReturnDataOutOfBounds    = errors.New("vm: return data out of bounds")
	ErrTraceLimitReached        = errors.New("vm: the number of logs reached the specified limit")
	ErrWriteProtection          = errors.New("vm: write protection")
	ErrContractRequire          = errors.New("vm: contract require fail")
	ErrContractAssert           = errors.New("vm: contract assert fail")
	ErrOutOfGas                 = errors.New("vm: out of gas")

	ErrOverFrame  = errors.New("engine: recursive overflow")
	ErrEmptyFrame = errors.New("engine: empty frame")
	ErrInitEngine = errors.New("engine: init failed")
)

type Error struct {
	err error
	fn  string
	app *APP
	eng *Engine
}

func (e Error) Error() string {
	return e.err.Error()
}
