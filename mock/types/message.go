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

package types

import (
	"math/big"
)

type Message struct {
	to         *Address
	from       Address
	nonce      uint64
	amount     *big.Int
	gasLimit   uint64
	gasPrice   *big.Int
	data       []byte
	checkNonce bool
}

func (m Message) From() Address      { return m.from }
func (m Message) To() *Address       { return m.to }
func (m Message) GasPrice() *big.Int { return m.gasPrice }
func (m Message) Value() *big.Int    { return m.amount }
func (m Message) Gas() uint64        { return m.gasLimit }
func (m Message) Nonce() uint64      { return m.nonce }
func (m Message) Data() []byte       { return m.data }
func (m Message) CheckNonce() bool   { return m.checkNonce }
