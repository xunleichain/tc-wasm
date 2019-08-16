// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package vm

import (
	"math/big"

	"github.com/xunleichain/tc-wasm/mock/types"
)

// ContractRef is a reference to the contract's backing object
type ContractRef interface {
	Address() types.Address
}

// AccountRef implements ContractRef.
//
// Account references are used during EVM initialisation and
// it's primary use is to fetch addresses. Removing this object
// proves difficult because of the cached jump destinations which
// are fetched from the parent contract (i.e. the caller), which
// is a ContractRef.
type AccountRef types.Address

// Address casts AccountRef to a Address
func (ar AccountRef) Address() types.Address { return (types.Address)(ar) }

// Contract represents an ethereum contract in the state database. It contains
// the the contract code, calling arguments. Contract implements ContractRef
type Contract struct {
	// CallerAddress is the result of the caller which initialised this
	// contract. However when the "call method" is delegated this value
	// needs to be initialised to that of the caller's caller.
	CallerAddress types.Address
	caller        ContractRef
	Self          ContractRef

	Code     []byte
	CodeHash types.Hash
	CodeAddr *types.Address
	Input    []byte

	Gas   uint64
	value *big.Int

	DelegateCall bool
	CreateCall   bool
}

// NewContract returns a new contract environment for the execution of EVM.
func NewContract(caller []byte, object []byte, value *big.Int, gas uint64) *Contract {
	callerRef := ContractRef(AccountRef(types.BytesToAddress(caller)))
	objectRef := AccountRef(types.BytesToAddress(object))
	c := &Contract{CallerAddress: callerRef.Address(), caller: callerRef, Self: objectRef}

	// Gas should be a pointer so it can safely be reduced through the run
	// This pointer will be off the state transition
	c.Gas = gas
	// ensures a value is set
	c.value = value
	c.CreateCall = false

	return c
}

// NewContract returns a new contract environment for the execution of call-contract.
func NewContractInner(caller ContractRef, object ContractRef, value *big.Int, gas uint64) *Contract {
	c := &Contract{CallerAddress: caller.Address(), caller: caller, Self: object}

	// Gas should be a pointer so it can safely be reduced through the run
	// This pointer will be off the state transition
	c.Gas = gas
	// ensures a value is set
	c.value = value
	c.CreateCall = false

	return c
}

// AsDelegate sets the contract to be a delegate call and returns the current
// contract (for chaining calls)
func (c *Contract) AsDelegate() *Contract {
	c.DelegateCall = true
	// NOTE: caller must, at all times be a contract. It should never happen
	// that caller is something other than a Contract.
	parent := c.caller.(*Contract)
	c.CallerAddress = parent.CallerAddress
	c.value = parent.value

	return c
}

// Caller returns the caller of the contract.
//
// Caller will recursively call caller when the contract is a delegate
// call, including that of caller's caller.
func (c *Contract) Caller() types.Address {
	return c.CallerAddress
}

// UseGas attempts the use gas and subtracts it and returns true on success
func (c *Contract) UseGas(gas uint64) (ok bool) {
	if c.Gas < gas {
		return false
	}
	c.Gas -= gas
	return true
}

// Address returns the contracts address
func (c *Contract) Address() types.Address {
	return c.Self.Address()
}

// Value returns the contracts value (sent to it from it's caller)
func (c *Contract) Value() *big.Int {
	return c.value
}

// SetCode sets the code to the contract
func (c *Contract) SetCode(hash types.Hash, code []byte) {
	c.Code = code
	c.CodeHash = hash
}

// SetCallCode sets the code of the contract and address of the backing data
// object
func (c *Contract) SetCallCode(addr []byte, hash []byte, code []byte) {
	c.Code = code
	c.CodeHash = types.BytesToHash(hash)
	codeAddr := types.BytesToAddress(addr)
	c.CodeAddr = &codeAddr
}
