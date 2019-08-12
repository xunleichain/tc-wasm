// Copyright 2014 The go-ethereum Authors
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

package state

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/xunleichain/tc-wasm/mock/types"
)

var emptyCodeHash = types.Keccak256(nil)

type Code []byte

func (c Code) String() string {
	return string(c) //strings.Join(Disassemble(self), " ")
}

type Storage map[types.Hash][]byte

func (s Storage) String() (str string) {
	for key, value := range s {
		str += fmt.Sprintf("%X : %X\n", key, value)
	}

	return
}

func (s Storage) Copy() Storage {
	cpy := make(Storage)
	for key, value := range s {
		cpy[key] = value
	}

	return cpy
}

// stateObject represents an Ethereum account which is being modified.
//
// The usage pattern is as follows:
// First you need to obtain a state object.
// Account values can be accessed and modified through the object.
// Finally, call CommitTrie to write the modified storage trie into a database.
type stateObject struct {
	address  types.Address
	addrHash types.Hash // hash of ethereum address of the account
	data     Account
	db       *StateDB

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error

	// Write caches.
	code Code // contract bytecode, which gets set when code is loaded

	originStorage Storage // Storage cache of original entries to dedup rewrites
	dirtyStorage  Storage // Storage entries that need to be flushed to disk

	// Cache flags.
	// When an object is marked suicided it will be delete from the trie
	// during the "update" phase of the state transition.
	dirtyCode bool // true if the code was updated
	suicided  bool
	deleted   bool
}

// empty returns whether the account is considered empty.
func (c *stateObject) empty() bool {
	return c.data.Nonce == 0 && c.data.Balance.Sign() == 0 && bytes.Equal(c.data.CodeHash, emptyCodeHash)
}

// Account is the Ethereum consensus representation of accounts.
// These objects are stored in the main account trie.
type Account struct {
	Nonce    uint64
	Credits  uint64
	Balance  *big.Int
	Tokens   map[types.Address]*big.Int
	Root     types.Hash // merkle(or kv) root of the storage trie
	CodeHash []byte
}

// newObject creates a state object.
func newObject(db *StateDB, address types.Address, data Account) *stateObject {
	if data.Balance == nil {
		data.Balance = new(big.Int)
	}
	if data.Tokens == nil {
		data.Tokens = make(map[types.Address]*big.Int)
	}
	if data.CodeHash == nil {
		data.CodeHash = emptyCodeHash
	}
	return &stateObject{
		db:            db,
		address:       address,
		addrHash:      types.Keccak256Hash(address[:]),
		data:          data,
		originStorage: make(Storage),
		dirtyStorage:  make(Storage),
	}
}

// setError remembers the first non-nil error it is called with.
func (c *stateObject) setError(err error) {
	if c.dbErr == nil {
		c.dbErr = err
	}
}

func (c *stateObject) markSuicided() {
	c.suicided = true
}

func (c *stateObject) touch() {
	c.db.journal.append(touchChange{
		account: &c.address,
	})
	if c.address == ripemd {
		// Explicitly put it in the dirty-cache, which is otherwise generated from
		// flattened journals.
		c.db.journal.dirty(c.address)
	}
}

// GetState returns a value in account storage.
func (c *stateObject) GetState(key types.Hash) []byte {
	value, dirty := c.dirtyStorage[key]
	if dirty {
		// If we have a dirty value for this state entry, return it
		return value
	}

	// Otherwise return the entry's original value
	return c.GetCommittedState(key)
}

// GetCommittedState retrieves a value from the committed account storage trie.
func (c *stateObject) GetCommittedState(key types.Hash) []byte {
	// If we have the original value cached, return that
	value, cached := c.originStorage[key]
	if cached {
		return value
	}
	return nil
}

// SetState updates a value in account storage.
func (c *stateObject) SetState(key types.Hash, value []byte) {
	// If the new value is the same as old, don't set
	prev := c.GetState(key)
	if bytes.Equal(prev, value) {
		return
	}
	// New value is different, update and journal the change
	c.db.journal.append(storageChange{
		account:  &c.address,
		key:      key,
		prevalue: prev,
	})
	c.setState(key, value)
}

func (c *stateObject) setState(key types.Hash, value []byte) {
	c.dirtyStorage[key] = value
}

// AddBalance removes amount from c's balance.
// It is used to add funds to the destination account of a transfer.
func (c *stateObject) AddBalance(amount *big.Int) {
	// EIP158: We must check emptiness for the objects such that the account
	// clearing (0,0,0 objects) can take effect.
	if amount.Sign() == 0 {
		if c.empty() {
			c.touch()
		}

		return
	}
	c.SetBalance(new(big.Int).Add(c.Balance(), amount))
}

// SubBalance removes amount from c's balance.
// It is used to remove funds from the origin account of a transfer.
func (c *stateObject) SubBalance(amount *big.Int) {
	if amount.Sign() == 0 {
		return
	}
	c.SetBalance(new(big.Int).Sub(c.Balance(), amount))
}

func (c *stateObject) SetBalance(amount *big.Int) {
	c.SetCredits(c.Credits() + 1)
	c.db.journal.append(balanceChange{
		account: &c.address,
		prev:    new(big.Int).Set(c.data.Balance),
	})
	c.setBalance(amount)
}

func (c *stateObject) setBalance(amount *big.Int) {
	c.data.Balance = amount
}

func (c *stateObject) AddTokenBalance(token types.Address, amount *big.Int) {
	// EIP158: We must check emptiness for the objects such that the account
	// clearing (0,0,0 objects) can take effect.
	if amount.Sign() == 0 {
		if c.empty() {
			c.touch()
		}

		return
	}
	c.SetTokenBalance(token, new(big.Int).Add(c.TokenBalance(token), amount))
}

func (c *stateObject) SubTokenBalance(token types.Address, amount *big.Int) {
	if amount.Sign() == 0 {
		return
	}
	c.SetTokenBalance(token, new(big.Int).Sub(c.TokenBalance(token), amount))
}

func (c *stateObject) SetTokenBalance(token types.Address, amount *big.Int) {
	if token == types.EmptyAddress {
		c.SetBalance(amount)
		return
	}
	if _, ok := c.data.Tokens[token]; !ok {
		c.data.Tokens[token] = Big0
	}
	c.SetCredits(c.Credits() + 1)
	c.db.journal.append(tokenBalanceChange{
		account: &c.address,
		token:   &token,
		prev:    new(big.Int).Set(c.data.Tokens[token]),
	})
	c.setTokenBalance(token, amount)
}

func (c *stateObject) setTokenBalance(token types.Address, amount *big.Int) {
	c.data.Tokens[token] = amount
}

// Return the gas back to the origin. Used by the Virtual machine or Closures
func (c *stateObject) ReturnGas(gas *big.Int) {}

func (c *stateObject) deepCopy(db *StateDB) *stateObject {
	stateObject := newObject(db, c.address, c.data)
	stateObject.code = c.code
	stateObject.dirtyStorage = c.dirtyStorage.Copy()
	stateObject.originStorage = c.originStorage.Copy()
	stateObject.suicided = c.suicided
	stateObject.dirtyCode = c.dirtyCode
	stateObject.deleted = c.deleted
	return stateObject
}

//
// Attribute accessors
//

// Returns the address of the contract/account
func (c *stateObject) Address() types.Address {
	return c.address
}

// Code returns the contract code associated with this object, if any.
func (c *stateObject) Code() []byte {
	if c.code != nil {
		return c.code
	}
	if bytes.Equal(c.CodeHash(), emptyCodeHash) {
		return nil
	}
	return nil
}

func (c *stateObject) SetCode(codeHash types.Hash, code []byte) {
	prevcode := c.Code()
	c.db.journal.append(codeChange{
		account:  &c.address,
		prevhash: c.CodeHash(),
		prevcode: prevcode,
	})
	c.setCode(codeHash, code)
}

func (c *stateObject) setCode(codeHash types.Hash, code []byte) {
	c.code = code
	c.data.CodeHash = codeHash[:]
	c.dirtyCode = true
}

func (c *stateObject) SetNonce(nonce uint64) {
	c.db.journal.append(nonceChange{
		account: &c.address,
		prev:    c.data.Nonce,
	})
	c.setNonce(nonce)
}

func (c *stateObject) setNonce(nonce uint64) {
	c.data.Nonce = nonce
}

func (c *stateObject) SetCredits(credits uint64) {
	c.db.journal.append(creditsChange{
		account: &c.address,
		prev:    c.data.Credits,
	})
	c.setCredits(credits)
}

func (c *stateObject) setCredits(credits uint64) {
	c.data.Credits = credits
}

func (c *stateObject) CodeHash() []byte {
	return c.data.CodeHash
}

func (c *stateObject) Balance() *big.Int {
	return c.data.Balance
}

func (c *stateObject) TokenBalance(token types.Address) *big.Int {
	if token == types.EmptyAddress {
		return c.data.Balance
	}
	if balance, ok := c.data.Tokens[token]; ok {
		return balance
	}
	return Big0
}

func (c *stateObject) TokenBalances() types.TokenValues {
	tv := make(types.TokenValues, 0, len(c.data.Tokens)+1)
	if c.data.Balance.Sign() > 0 {
		tv = append(tv, types.TokenValue{
			TokenAddr: types.EmptyAddress,
			Value:     big.NewInt(0).Set(c.data.Balance),
		})
	}
	for addr, val := range c.data.Tokens {
		if val.Sign() > 0 {
			tv = append(tv, types.TokenValue{
				TokenAddr: addr,
				Value:     big.NewInt(0).Set(val),
			})
		}
	}
	return tv
}

func (c *stateObject) Nonce() uint64 {
	return c.data.Nonce
}

func (c *stateObject) Credits() uint64 {
	return c.data.Credits
}

func (c *stateObject) StorageRoot() types.Hash {
	return c.data.Root
}

// Never called, but must be present to allow stateObject to be used
// as a vm.Account interface that also satisfies the vm.ContractRef
// interface. Interfaces are awesome.
func (c *stateObject) Value() *big.Int {
	panic("Value on stateObject should never be called")
}
