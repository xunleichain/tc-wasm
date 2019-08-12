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

// Package state provides a caching layer atop the Ethereum state trie.
package state

import (
	"bytes"
	"fmt"
	"math/big"
	"sort"

	"github.com/xunleichain/tc-wasm/mock/log"
	"github.com/xunleichain/tc-wasm/mock/types"
)

type revision struct {
	id           int
	journalIndex int
}

var (
	_ types.StateDB = &StateDB{}

	Big0 = big.NewInt(0)
)

// StateDBs within the ethereum protocol are used to store anything
// within the merkle trie. StateDBs take care of caching and storing
// nested states. It's the general query interface to retrieve:
// * Contracts
// * Accounts
type StateDB struct {
	// This map holds 'live' objects, which will get modified while processing a state transition.
	stateObjects      map[types.Address]*stateObject
	stateObjectsDirty map[types.Address]struct{}

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error

	// The refund counter, also used by state transitioning.
	refund uint64

	thash, bhash types.Hash
	txIndex      int
	logs         map[types.Hash][]*types.Log
	logSize      uint

	preimages map[types.Hash][]byte

	// Journal of state modifications. This is the backbone of
	// Snapshot and RevertToSnapshot.
	journal        *journal
	validRevisions []revision
	nextRevisionId int
}

// Create a new state from a given trie.
func New() (*StateDB, error) {
	return &StateDB{
		stateObjects:      make(map[types.Address]*stateObject),
		stateObjectsDirty: make(map[types.Address]struct{}),
		logs:              make(map[types.Hash][]*types.Log),
		preimages:         make(map[types.Hash][]byte),
		journal:           newJournal(),
	}, nil
}

// setError remembers the first non-nil error it is called with.
func (s *StateDB) setError(err error) {
	if s.dbErr == nil {
		s.dbErr = err
	}
}

func (s *StateDB) Error() error {
	return s.dbErr
}

// Reset clears out all ephemeral state objects from the state db, but keeps
// the underlying state trie to avoid reloading data for the next operations.
func (s *StateDB) Reset(root types.Hash) error {
	s.stateObjects = make(map[types.Address]*stateObject)
	s.stateObjectsDirty = make(map[types.Address]struct{})
	s.thash = types.EmptyHash
	s.bhash = types.EmptyHash
	s.txIndex = 0
	s.logs = make(map[types.Hash][]*types.Log)
	s.logSize = 0
	s.preimages = make(map[types.Hash][]byte)
	s.clearJournalAndRefund()
	return nil
}

func (s *StateDB) AddLog(log *types.Log) {
	s.journal.append(addLogChange{txhash: s.thash})

	log.TxHash = s.thash
	log.BlockHash = s.bhash
	log.TxIndex = uint(s.txIndex)
	log.Index = s.logSize
	s.logs[s.thash] = append(s.logs[s.thash], log)
	s.logSize++
}

func (s *StateDB) GetLogs(hash types.Hash) []*types.Log {
	return s.logs[hash]
}

func (s *StateDB) Logs() []*types.Log {
	var logs []*types.Log
	for _, lgs := range s.logs {
		logs = append(logs, lgs...)
	}
	return logs
}

// AddPreimage records a SHA3 preimage seen by the VM.
func (s *StateDB) AddPreimage(hash types.Hash, preimage []byte) {
	if _, ok := s.preimages[hash]; !ok {
		s.journal.append(addPreimageChange{hash: hash})
		pi := make([]byte, len(preimage))
		copy(pi, preimage)
		s.preimages[hash] = pi
	}
}

// Preimages returns a list of SHA3 preimages that have been submitted.
func (s *StateDB) Preimages() map[types.Hash][]byte {
	return s.preimages
}

// AddRefund adds gas to the refund counter
func (s *StateDB) AddRefund(gas uint64) {
	s.journal.append(refundChange{prev: s.refund})
	s.refund += gas
}

// SubRefund removes gas from the refund counter.
// This method will panic if the refund counter goes below zero
func (s *StateDB) SubRefund(gas uint64) {
	s.journal.append(refundChange{prev: s.refund})
	if gas > s.refund {
		panic("Refund counter below zero")
	}
	s.refund -= gas
}

// Exist reports whether the given account address exists in the state.
// Notably this also returns true for suicided accounts.
func (s *StateDB) Exist(addr types.Address) bool {
	return s.getStateObject(addr) != nil
}

// Empty returns whether the state object is either non-existent
// or empty according to the EIP161 specification (balance = nonce = code = 0)
func (s *StateDB) Empty(addr types.Address) bool {
	so := s.getStateObject(addr)
	return so == nil || so.empty()
}

// Retrieve the balance from the given address or 0 if object not found
func (s *StateDB) GetBalance(addr types.Address) *big.Int {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Balance()
	}
	return Big0
}

func (s *StateDB) GetTokenBalance(addr types.Address, token types.Address) *big.Int {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.TokenBalance(token)
	}
	return Big0
}

func (s *StateDB) GetTokenBalances(addr types.Address) types.TokenValues {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.TokenBalances()
	}
	return nil
}

func (s *StateDB) GetNonce(addr types.Address) uint64 {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Nonce()
	}
	return 0
}

func (s *StateDB) GetCredits(addr types.Address) uint64 {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Credits()
	}
	return 0
}

func (s *StateDB) GetCode(addr types.Address) []byte {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Code()
	}
	return nil
}

func (s *StateDB) GetCodeSize(addr types.Address) int {
	stateObject := s.getStateObject(addr)
	if stateObject == nil {
		return 0
	}
	if stateObject.code != nil {
		return len(stateObject.code)
	}
	return 0
}

func (s *StateDB) IsContract(addr types.Address) bool {
	stateObject := s.getStateObject(addr)
	if stateObject == nil {
		return false
	}
	if bytes.Equal(stateObject.CodeHash(), emptyCodeHash) {
		return false
	}
	return true
}

func (s *StateDB) GetCodeHash(addr types.Address) types.Hash {
	stateObject := s.getStateObject(addr)
	if stateObject == nil {
		return types.EmptyHash
	}
	return types.BytesToHash(stateObject.CodeHash())
}

// GetState retrieves a value from the given account's storage trie.
func (s *StateDB) GetState(addr types.Address, hash types.Hash) []byte {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.GetState(hash)
	}
	return nil
}

// GetCommittedState retrieves a value from the given account's committed storage trie.
func (s *StateDB) GetCommittedState(addr types.Address, hash types.Hash) []byte {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.GetCommittedState(hash)
	}
	return nil
}

func (s *StateDB) GetStorageRoot(addr types.Address) types.Hash {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.StorageRoot()
	}
	return types.EmptyHash
}

func (s *StateDB) HasSuicided(addr types.Address) bool {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.suicided
	}
	return false
}

/*
 * SETTERS
 */

// AddBalance adds amount to the account associated with addr.
func (s *StateDB) AddBalance(addr types.Address, amount *big.Int) {
	//TODO amount == 0 return
	stateObject := s.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.AddBalance(amount)
	}
}

// SubBalance subtracts amount from the account associated with addr.
func (s *StateDB) SubBalance(addr types.Address, amount *big.Int) {
	stateObject := s.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SubBalance(amount)
	}
}

func (s *StateDB) SetBalance(addr types.Address, amount *big.Int) {
	stateObject := s.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetBalance(amount)
	}
}

func (s *StateDB) SetNonce(addr types.Address, nonce uint64) {
	stateObject := s.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetNonce(nonce)
	}
}

func (s *StateDB) SetCredits(addr types.Address, credits uint64) {
	stateObject := s.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetCredits(credits)
	}
}

func (s *StateDB) SetCode(addr types.Address, code []byte) {
	stateObject := s.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetCode(types.Keccak256Hash(code), code)
	}
}

func (s *StateDB) SetState(addr types.Address, key types.Hash, value []byte) {
	stateObject := s.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetState(key, value)
	}
}

func (s *StateDB) AddTokenBalance(addr types.Address, token types.Address, amount *big.Int) {
	stateObject := s.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.AddTokenBalance(token, amount)
	}
}

// SubBalance subtracts amount from the account associated with addr.
func (s *StateDB) SubTokenBalance(addr types.Address, token types.Address, amount *big.Int) {
	stateObject := s.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SubTokenBalance(token, amount)
	}
}

func (s *StateDB) SetTokenBalance(addr types.Address, token types.Address, amount *big.Int) {
	stateObject := s.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetTokenBalance(token, amount)
	}
}

// Suicide marks the given account as suicided.
// This clears the account balance.
//
// The account's state object is still available until the state is committed,
// getStateObject will return a non-nil account after Suicide.
func (s *StateDB) Suicide(addr types.Address) bool {
	stateObject := s.getStateObject(addr)
	if stateObject == nil {
		return false
	}

	tv := stateObject.TokenBalances()
	for i := 0; i < len(tv); i++ {
		log.Debug("Suicide tokens", "token", tv[i].TokenAddr.String(), "value", tv[i].Value)
	}

	s.journal.append(suicideChange{
		account:     &addr,
		prev:        stateObject.suicided,
		prevbalance: tv,
	})
	stateObject.markSuicided()
	stateObject.data.Balance = new(big.Int)
	stateObject.data.Tokens = make(map[types.Address]*big.Int)

	return true
}

//
// Setting, updating & deleting state object methods.
//

// Retrieve a state object given by the address. Returns nil if not found.
func (s *StateDB) getStateObject(addr types.Address) (stateObject *stateObject) {
	// Prefer 'live' objects.
	if obj := s.stateObjects[addr]; obj != nil {
		if obj.deleted {
			return nil
		}
		return obj
	}
	return nil
}

func (s *StateDB) setStateObject(object *stateObject) {
	s.stateObjects[object.Address()] = object
}

// Retrieve a state object or create a new state object if nil.
func (s *StateDB) GetOrNewStateObject(addr types.Address) *stateObject {
	stateObject := s.getStateObject(addr)
	if stateObject == nil || stateObject.deleted {
		stateObject, _ = s.createObject(addr)
	}
	return stateObject
}

// createObject creates a new state object. If there is an existing account with
// the given address, it is overwritten and returned as the second return value.
func (s *StateDB) createObject(addr types.Address) (newobj, prev *stateObject) {
	prev = s.getStateObject(addr)
	newobj = newObject(s, addr, Account{Credits: 1})
	newobj.setNonce(0) // sets the object to dirty
	if prev == nil {
		s.journal.append(createObjectChange{account: &addr})
	} else {
		s.journal.append(resetObjectChange{prev: prev})
	}
	s.setStateObject(newobj)
	return newobj, prev
}

// CreateAccount explicitly creates a state object. If a state object with the address
// already exists the balance is carried over to the new account.
//
// CreateAccount is called during the EVM CREATE operation. The situation might arise that
// a contract does the following:
//
//   1. sends funds to sha(account ++ (nonce + 1))
//   2. tx_create(sha(account ++ nonce)) (note that this gets the address of 1)
//
// Carrying over the balance ensures that Ether doesn't disappear.
func (s *StateDB) CreateAccount(addr types.Address) {
	newObj, prev := s.createObject(addr)
	if prev != nil {
		newObj.setBalance(prev.data.Balance)
	}
}

func (s *StateDB) GetAccount(address types.Address) *Account {
	so := s.getStateObject(address)
	if so == nil {
		return nil
	}
	return &so.data
}

func (s *StateDB) ForEachStorage(addr types.Address, cb func(key types.Hash, value []byte) bool) {
	so := s.getStateObject(addr)
	if so == nil {
		return
	}
	// TODO: cb dirty first
	return
}

// Copy creates a deep, independent copy of the state.
// Snapshots of the copied state cannot be applied to the copy.
func (s *StateDB) Copy() *StateDB {
	// Copy all the basic fields, initialize the memory ones
	state := &StateDB{
		stateObjects:      make(map[types.Address]*stateObject, len(s.journal.dirties)),
		stateObjectsDirty: make(map[types.Address]struct{}, len(s.journal.dirties)),
		refund:            s.refund,
		logs:              make(map[types.Hash][]*types.Log, len(s.logs)),
		logSize:           s.logSize,
		preimages:         make(map[types.Hash][]byte),
		journal:           newJournal(),
	}
	// Copy the dirty states, logs, and preimages
	for addr := range s.journal.dirties {
		// As documented [here](https://github.com/ethereum/go-ethereum/pull/16485#issuecomment-380438527),
		// and in the Finalise-method, there is a case where an object is in the journal but not
		// in the stateObjects: OOG after touch on ripeMD prior to Byzantium. Thus, we need to check for
		// nil
		if object, exist := s.stateObjects[addr]; exist {
			state.stateObjects[addr] = object.deepCopy(state)
			state.stateObjectsDirty[addr] = struct{}{}
		}
	}
	// Above, we don't copy the actual journal. This means that if the copy is copied, the
	// loop above will be a no-op, since the copy's journal is empty.
	// Thus, here we iterate over stateObjects, to enable copies of copies
	for addr := range s.stateObjectsDirty {
		if _, exist := state.stateObjects[addr]; !exist {
			state.stateObjects[addr] = s.stateObjects[addr].deepCopy(state)
			state.stateObjectsDirty[addr] = struct{}{}
		}
	}

	for hash, logs := range s.logs {
		cpy := make([]*types.Log, len(logs))
		for i, l := range logs {
			cpy[i] = new(types.Log)
			*cpy[i] = *l
		}
		state.logs[hash] = cpy
	}
	for hash, preimage := range s.preimages {
		state.preimages[hash] = preimage
	}
	return state
}

// Snapshot returns an identifier for the current revision of the state.
func (s *StateDB) Snapshot() int {
	id := s.nextRevisionId
	s.nextRevisionId++
	s.validRevisions = append(s.validRevisions, revision{id, s.journal.length()})
	return id
}

// RevertToSnapshot reverts all state changes made since the given revision.
func (s *StateDB) RevertToSnapshot(revid int) {
	// Find the snapshot in the stack of valid snapshots.
	idx := sort.Search(len(s.validRevisions), func(i int) bool {
		return s.validRevisions[i].id >= revid
	})
	if idx == len(s.validRevisions) || s.validRevisions[idx].id != revid {
		panic(fmt.Errorf("revision id %v cannot be reverted", revid))
	}
	snapshot := s.validRevisions[idx].journalIndex

	// Replay the journal to undo changes and remove invalidated snapshots
	s.journal.revert(s, snapshot)
	s.validRevisions = s.validRevisions[:idx]
}

// GetRefund returns the current value of the refund counter.
func (s *StateDB) GetRefund() uint64 {
	return s.refund
}

// Prepare sets the current transaction hash and index and block hash which is
// used when the EVM emits new state logs.
func (s *StateDB) Prepare(thash, bhash types.Hash, ti int) {
	s.thash = thash
	s.bhash = bhash
	s.txIndex = ti
}

// TxHash get the current Transaction Hash.
func (s *StateDB) TxHash() types.Hash {
	return s.thash
}

func (s *StateDB) clearJournalAndRefund() {
	s.journal = newJournal()
	s.validRevisions = s.validRevisions[:0]
	s.refund = 0
}
