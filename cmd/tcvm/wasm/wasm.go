package wasm

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/xunleichain/tc-wasm/mock/log"
	"github.com/xunleichain/tc-wasm/mock/types"
	"github.com/xunleichain/tc-wasm/vm"
)

var (
	defaultDifficulty = big.NewInt(10000000)
)

// emptyCodeHash is used by create to ensure deployment is disallowed to already
// deployed contract addresses (relevant after the account abstraction).
var emptyCodeHash = types.Keccak256Hash(nil)

// ChainContext supports retrieving headers from the
// current blockchain to be used during transaction processing.
type ChainContext interface {
	// GetHeader returns the hash corresponding to their hash.
	GetHeader(uint64) *types.Header
}

type (
	// CanTransferFunc is the signature of a transfer guard function
	CanTransferFunc func(types.StateDB, types.Address, types.Address, *big.Int) bool
	// TransferFunc is the signature of a transfer function
	TransferFunc func(types.StateDB, types.Address, types.Address, types.Address, *big.Int)
	// GetHashFunc returns the nth block hash in the blockchain
	// and is used by the BLOCKHASH EVM op code.
	GetHashFunc func(uint64) types.Hash
)

// Context provides the EVM with auxiliary information. Once provided
// it shouldn't be modified.
type Context struct {
	// CanTransfer returns whether the account contains
	// sufficient ether to transfer the value
	CanTransfer CanTransferFunc
	// Transfer transfers ether from one account to the other
	Transfer TransferFunc
	// GetHash returns the hash corresponding to n
	GetHash GetHashFunc

	// Message information
	Token    types.Address // Provides the tx token type
	Origin   types.Address // Provides information for ORIGIN
	GasPrice *big.Int      // Provides information for GASPRICE

	// Block information
	Coinbase    types.Address // Provides information for COINBASE
	GasLimit    uint64        // Provides information for GASLIMIT
	BlockNumber *big.Int      // Provides information for NUMBER
	Time        *big.Int      // Provides information for TIME
	Difficulty  *big.Int      // Provides information for DIFFICULTY

	//Use tx nonce to compute contract address in version2
	IsVersion2  bool
	WasmGasRate uint64
	Nonce       uint64
}

// NewWASMContext creates a new context for use in the WASM.
func NewWASMContext(header *types.Header, chain ChainContext, author *types.Address, gasRate uint64) Context {
	// If we don't have an explicit author (i.e. not mining), extract from the header
	var beneficiary types.Address
	if author == nil {
		beneficiary = header.Coinbase // Ignore error, we're past header validation
	} else {
		beneficiary = *author
	}

	ctx := Context{
		CanTransfer: CanTransfer,
		Transfer:    Transfer,
		GetHash:     GetHashFn(header, chain),
		Coinbase:    beneficiary,
		BlockNumber: new(big.Int).SetUint64(header.Height),
		Time:        new(big.Int).SetUint64(header.Time),
		Difficulty:  new(big.Int).Set(defaultDifficulty),
		GasLimit:    header.GasLimit,
		WasmGasRate: gasRate,
		Nonce:       0,
		IsVersion2:  false,
	}

	if ctx.Time.Cmp(vm.TsVersion2Sec) >= 0 {
		ctx.IsVersion2 = true
	}

	return ctx
}

// GetHashFn returns a GetHashFunc which retrieves header hashes by number
func GetHashFn(ref *types.Header, chain ChainContext) func(n uint64) types.Hash {
	var cache map[uint64]types.Hash

	return func(n uint64) types.Hash {
		// If there's no hash cache yet, make one
		if cache == nil {
			cache = map[uint64]types.Hash{
				ref.Height - 1: ref.ParentHash,
			}
		}
		// Try to fulfill the request from the cache
		if hash, ok := cache[n]; ok {
			return hash
		}
		// Not cached, iterate the blocks and cache the hashes
		for header := chain.GetHeader(ref.Height - 1); header != nil; header = chain.GetHeader(header.Height - 1) {
			cache[header.Height-1] = header.ParentHash
			if n == header.Height-1 {
				return header.ParentHash
			}

			if header.Height == 0 {
				break
			}
		}
		return types.EmptyHash
	}
}

// CanTransfer checks wether there are enough funds in the address' account to make a transfer.
// This does not take the necessary gas in to account to make the transfer valid.
func CanTransfer(db types.StateDB, addr, token types.Address, amount *big.Int) bool {
	return db.GetTokenBalance(addr, token).Cmp(amount) >= 0
}

// Transfer subtracts amount from sender and adds amount to recipient using the given Db
func Transfer(db types.StateDB, sender, recipient, token types.Address, amount *big.Int) {
	db.SubTokenBalance(sender, token, amount)
	db.AddTokenBalance(recipient, token, amount)
}

// run runs the given contract and takes care of running precompiles with a fallback to the byte code interpreter.
func run(wasm *WASM, c types.Contract, input []byte) ([]byte, uint64, error) {
	contract := c.(*vm.Contract)
	localMaxGas := new(big.Int).Mul(new(big.Int).SetUint64(contract.Gas), new(big.Int).SetUint64(wasm.WasmGasRate)).Uint64()

	addr := contract.CodeAddr

	eng := vm.NewEngine(contract, localMaxGas, wasm.StateDB, log.With("mod", "wasm"))
	eng.SetTrace(false)
	app, err := eng.NewApp(addr.String(), nil, false)
	if err != nil {
		if err == vm.ErrContractNoCode {
			return nil, contract.Gas, nil
		}
		log.Error("WASM eng.NewApp", "err", err, "contract", addr.String())
		return nil, contract.Gas, fmt.Errorf("WASM eng.NewApp,err:%v", err)
	}

	fnIndex := app.GetExportFunction(vm.APPEntry)
	if fnIndex < 0 {
		return []byte(""), contract.Gas, fmt.Errorf("GetExportFunction(APPEntry) fail")
	}

	ret, err := eng.Run(app, input)
	gasused, modgas := new(big.Int).DivMod(new(big.Int).SetUint64(eng.GasUsed()), new(big.Int).SetUint64(wasm.WasmGasRate), big.NewInt(0))
	subModGas := uint64(0)
	if modgas.Uint64() > 0 {
		subModGas = uint64(1)
	}

	gas := contract.Gas - gasused.Uint64() - subModGas

	if err != nil {
		log.Error("WASM eng.Run ret:", "ret", ret, "gas", gas, "err", err)
		return nil, gas, err
	}

	// @Todo: Bugs here.
	retData, err := app.VM.VMemory().GetString(ret)
	log.Debug("WASM eng.Run ret:", "ret", ret, "retData", string(retData), "gas", gas, "eng.GasUsed()", eng.GasUsed(), "err", err)
	if err != nil {
		return nil, gas, err
	}

	return []byte(retData), gas, err
}

type Config struct {
}

type WASM struct {
	// Context provides auxiliary blockchain related information
	Context
	// StateDB gives access to the underlying state
	StateDB types.StateDB

	env *vm.EnvTable
	eng *vm.Engine
	app *vm.APP
}

// NewWASM returns a new WASM. The returned WASM is not thread safe and should
// only ever be used *once*.
func NewWASM(c types.Context, statedb types.StateDB, vmc types.VmConfig) *WASM {
	ctx := c.(Context)

	return &WASM{
		Context: ctx,
		StateDB: statedb,
	}
}

// reset
//func (wasm *WASM) Reset(origin types.Address, gasPrice *big.Int, nonce uint64) {
func (wasm *WASM) Reset(msg types.Message) {
	wasm.Context.Origin = msg.From()                         //origin
	wasm.Context.GasPrice = new(big.Int).Set(msg.GasPrice()) //gasPrice
	wasm.Context.Token = types.EmptyAddress
	wasm.Context.Nonce = msg.Nonce() //nonce
}

func (wasm *WASM) GetCode(bz []byte) []byte {
	return wasm.StateDB.GetCode(types.BytesToAddress(bz))
}

// Cancel cancels any running WASM operation. This may be called concurrently and
// it's safe to be called multiple times.
func (wasm *WASM) Cancel() {}

// Call executes the contract associated with the addr with the given input as
// parameters. It also handles any necessary value transfer required and takes
// the necessary steps to create accounts and reverses the state in case of an
// execution error or failed value transfer.
func (wasm *WASM) Call(c types.ContractRef, addr, token types.Address, input []byte, gas uint64, value *big.Int) (ret []byte, leftOverGas uint64, err error) {
	caller := c.(vm.ContractRef)

	// Fail if we're trying to transfer more than the available balance
	if !wasm.Context.CanTransfer(wasm.StateDB, caller.Address(), token, value) {
		return nil, gas, vm.ErrInsufficientBalance
	}

	var (
		to       = vm.AccountRef(addr)
		snapshot = wasm.StateDB.Snapshot()
	)
	if !wasm.StateDB.Exist(addr) {
		wasm.StateDB.CreateAccount(addr)
	}
	wasm.Transfer(wasm.StateDB, caller.Address(), to.Address(), token, value)

	// Initialise a new contract and set the code that is to be used by the WASM.
	// The contract is a scoped environment for this execution context only.
	contract := vm.NewContract(caller.Address().Bytes(), to.Address().Bytes(), value, gas)
	contract.SetCallCode(addr.Bytes(), wasm.StateDB.GetCodeHash(addr).Bytes(), wasm.StateDB.GetCode(addr))
	contract.Input = input

	ret, leftOverGas, err = run(wasm, contract, input)
	contract.Gas = leftOverGas
	// When an error was returned by the WASM or when setting the creation code
	// above we revert to the snapshot and consume any gas remaining. Additionally
	// when we're in homestead this also counts for code storage gas errors.
	if err != nil {
		wasm.StateDB.RevertToSnapshot(snapshot)
		if err != vm.ErrExecutionReverted {
			contract.UseGas(contract.Gas)
		}
	}
	return ret, contract.Gas, err
}

// CallCode executes the contract associated with the addr with the given input
// as parameters. It also handles any necessary value transfer required and takes
// the necessary steps to create accounts and reverses the state in case of an
// execution error or failed value transfer.
//
// CallCode differs from Call in the sense that it executes the given address'
// code with the caller as context.
func (wasm *WASM) CallCode(c types.ContractRef, addr types.Address, input []byte, gas uint64, value *big.Int) (ret []byte, leftOverGas uint64, err error) {
	caller := c.(vm.ContractRef)

	// Fail if we're trying to transfer more than the available balance
	if !wasm.CanTransfer(wasm.StateDB, caller.Address(), types.EmptyAddress, value) {
		return nil, gas, vm.ErrInsufficientBalance
	}

	var (
		snapshot = wasm.StateDB.Snapshot()
		to       = vm.AccountRef(caller.Address())
	)
	// initialise a new contract and set the code that is to be used by the
	// WASM. The contract is a scoped environment for this execution context
	// only.
	contract := vm.NewContract(caller.Address().Bytes(), to.Address().Bytes(), value, gas)
	contract.SetCallCode(addr.Bytes(), wasm.StateDB.GetCodeHash(addr).Bytes(), wasm.StateDB.GetCode(addr))
	contract.Input = input

	ret, leftOverGas, err = run(wasm, contract, input)
	contract.Gas = leftOverGas
	if err != nil {
		wasm.StateDB.RevertToSnapshot(snapshot)
		if err != vm.ErrExecutionReverted {
			contract.UseGas(contract.Gas)
		}
	}
	return ret, contract.Gas, err
}

// DelegateCall executes the contract associated with the addr with the given input
// as parameters. It reverses the state in case of an execution error.
//
// DelegateCall differs from CallCode in the sense that it executes the given address'
// code with the caller as context and the caller is set to the caller of the caller.
func (wasm *WASM) DelegateCall(c types.ContractRef, addr types.Address, input []byte, gas uint64) (ret []byte, leftOverGas uint64, err error) {
	caller := c.(vm.ContractRef)

	var (
		snapshot = wasm.StateDB.Snapshot()
		to       = vm.AccountRef(caller.Address())
	)

	// Initialise a new contract and make initialise the delegate values
	contract := vm.NewContract(caller.Address().Bytes(), to.Address().Bytes(), nil, gas).AsDelegate()
	contract.SetCallCode(addr.Bytes(), wasm.StateDB.GetCodeHash(addr).Bytes(), wasm.StateDB.GetCode(addr))
	contract.Input = input

	ret, leftOverGas, err = run(wasm, contract, input)
	contract.Gas = leftOverGas
	if err != nil {
		wasm.StateDB.RevertToSnapshot(snapshot)
		if err != vm.ErrExecutionReverted {
			contract.UseGas(contract.Gas)
		}
	}
	return ret, contract.Gas, err
}

// StaticCall executes the contract associated with the addr with the given input
// as parameters while disallowing any modifications to the state during the call.
// Opcodes that attempt to perform such modifications will result in exceptions
// instead of performing the modifications.
func (wasm *WASM) StaticCall(c types.ContractRef, addr types.Address, input []byte, gas uint64) (ret []byte, leftOverGas uint64, err error) {
	caller := c.(vm.ContractRef)

	// Make sure the readonly is only set if we aren't in readonly yet
	// this makes also sure that the readonly flag isn't removed for
	// child calls.
	// if !wasm.interpreter.readOnly {
	// 	wasm.interpreter.readOnly = true
	// 	defer func() { wasm.interpreter.readOnly = false }()
	// }

	var (
		to       = vm.AccountRef(addr)
		snapshot = wasm.StateDB.Snapshot()
	)
	// Initialise a new contract and set the code that is to be used by the
	// WASM. The contract is a scoped environment for this execution context
	// only.
	contract := vm.NewContract(caller.Address().Bytes(), to.Address().Bytes(), new(big.Int), gas)
	contract.SetCallCode(addr.Bytes(), wasm.StateDB.GetCodeHash(addr).Bytes(), wasm.StateDB.GetCode(addr))
	contract.Input = input

	// When an error was returned by the WASM or when setting the creation code
	// above we revert to the snapshot and consume any gas remaining. Additionally
	// when we're in Homestead this also counts for code storage gas errors.
	ret, leftOverGas, err = run(wasm, contract, input)
	contract.Gas = leftOverGas
	if err != nil {
		wasm.StateDB.RevertToSnapshot(snapshot)
		if err != vm.ErrExecutionReverted {
			contract.UseGas(contract.Gas)
		}
	}
	return ret, contract.Gas, err
}

// Create creates a new contract using code as deployment code.
func (wasm *WASM) Create(c types.ContractRef, data []byte, gas uint64, value *big.Int) (ret []byte, contractAddr types.Address, leftOverGas uint64, err error) {
	caller := c.(vm.ContractRef)

	if !wasm.CanTransfer(wasm.StateDB, caller.Address(), types.EmptyAddress, value) {
		return nil, types.EmptyAddress, gas, vm.ErrInsufficientBalance
	}

	// parse Constructor's arguments && bytecode
	input, code, err := vm.ParseInitArgsAndCode(data)
	if err != nil {
		log.Error("WASM Create: parse InitArgs Length fail", "err", err)
		return nil, types.EmptyAddress, gas, fmt.Errorf("Invalid InitArgs Length for Contract Init Function")
	}

	// Ensure there's no existing contract already at the designated address
	nonce := wasm.StateDB.GetNonce(caller.Address())
	wasm.StateDB.SetNonce(caller.Address(), nonce+1)

	if wasm.IsVersion2 {
		contractAddr = types.CreateAddress(caller.Address(), wasm.Nonce, code)
	} else {
		contractAddr = types.CreateAddress(caller.Address(), nonce, code)
	}

	contractHash := wasm.StateDB.GetCodeHash(contractAddr)
	if wasm.StateDB.GetNonce(contractAddr) != 0 || (contractHash != types.EmptyHash && contractHash != emptyCodeHash) {
		return nil, types.EmptyAddress, 0, vm.ErrContractAddressCollision
	}
	// Create a new account on the state
	snapshot := wasm.StateDB.Snapshot()
	wasm.StateDB.CreateAccount(contractAddr)
	wasm.StateDB.SetNonce(contractAddr, 1)
	wasm.StateDB.SetCode(contractAddr, code)

	encodeinput := hex.EncodeToString(input)
	strInput, _ := hex.DecodeString(encodeinput)

	wasm.Transfer(wasm.StateDB, caller.Address(), contractAddr, types.EmptyAddress, value)

	// initialise a new contract and set the code that is to be used by the
	// WASM. The contract is a scoped environment for this execution context
	// only.
	contract := vm.NewContract(caller.Address().Bytes(), contractAddr.Bytes(), value, gas)
	contract.SetCallCode(contractAddr.Bytes(), types.Keccak256Hash(code).Bytes(), code)
	contract.Input = []byte(strInput)
	contract.CreateCall = true

	// TODO :wasm not found code ,return err,create fail,
	ret, leftOverGas, err = run(wasm, contract, contract.Input)

	ret = code
	contract.Gas = leftOverGas

	// check whether the max code size has been exceeded
	maxCodeSizeExceeded := len(ret) > vm.MaxCodeSize
	// if the contract creation ran successfully and no errors were returned
	// calculate the gas required to store the code. If the code could not
	// be stored due to not enough gas set an error and let it be handled
	// by the error checking condition below.
	if err == nil && !maxCodeSizeExceeded {
		createDataGas := uint64(len(ret)) * vm.CreateDataGas / wasm.WasmGasRate
		contract.Gas = leftOverGas
		if contract.UseGas(createDataGas) {
			wasm.StateDB.SetCode(contractAddr, ret)
		} else {
			err = vm.ErrCodeStoreOutOfGas
		}
	}

	// When an error was returned by the WASM or when setting the creation code
	// above we revert to the snapshot and consume any gas remaining. Additionally
	// when we're in homestead this also counts for code storage gas errors.
	if maxCodeSizeExceeded || err != nil {
		wasm.StateDB.RevertToSnapshot(snapshot)
		if err != vm.ErrExecutionReverted {
			contract.UseGas(contract.Gas)
		}
	}
	// Assign err if contract code size exceeds the max while the err is still empty.
	if maxCodeSizeExceeded && err == nil {
		err = vm.ErrMaxCodeSizeExceeded
	}
	return ret, contractAddr, contract.Gas, err
}

// Interpreter returns the WASM interpreter
// func (wasm *WASM) Interpreter() *Interpreter { return wasm.interpreter }

func (wasm *WASM) Upgrade(caller types.ContractRef, contractAddr types.Address, code []byte) {
	wasm.StateDB.SetCode(contractAddr, code)
	wasm.StateDB.SetNonce(wasm.Context.Origin, wasm.Context.Nonce+1)
	vm.AppCache.Delete(contractAddr.String())
}

//Token
func (wasm *WASM) SetToken(token types.Address) {
	wasm.Token = token
}

// Coinbase
func (wasm *WASM) GetCoinbase() types.Address {
	return wasm.Coinbase
}

func (wasm *WASM) GetBlockNumber() *big.Int {
	return wasm.BlockNumber
}

//Time
func (wasm *WASM) GetTime() *big.Int {
	return wasm.Time
}

func (wasm *WASM) ISVersion2() bool {
	return wasm.IsVersion2
}

func (wasm *WASM) GasRate() uint64 {
	return wasm.WasmGasRate
}

//StateDB
func (wasm *WASM) GetStateDB() types.StateDB {
	return wasm.StateDB
}
