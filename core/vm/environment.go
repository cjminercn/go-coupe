// Copyright 2014 The go-coupe Authors
// This file is part of the go-coupe library.
//
// The go-coupe library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-coupe library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-coupe library. If not, see <http://www.gnu.org/licenses/>.

package vm

import (
	"math/big"

	"github.com/cjminercn/go-coupe/common"
)

// Environment is is required by the virtual machine to get information from
// it's own isolated environment.

// Environment is an EVM requirement and helper which allows access to outside
// information such as states.
type Environment interface {
	// The state database
	Db() Database
	// Creates a restorable snapshot
	MakeSnapshot() Database
	// Set database to previous snapshot
	SetSnapshot(Database)
	// Address of the original invoker (first occurance of the VM invoker)
	Origin() common.Address
	// The block number this VM is invoken on
	BlockNumber() *big.Int
	// The n'th hash ago from this block number
	GetHash(uint64) common.Hash
	// The handler's address
	Coinbase() common.Address
	// The current time (block time)
	Time() *big.Int
	// Difficulty set on the current block
	Difficulty() *big.Int
	// The gas limit of the block
	GasLimit() *big.Int
	// Determines whether it's possible to transact
	CanTransfer(from common.Address, balance *big.Int) bool
	// Transfers amount from one account to the other
	Transfer(from, to Account, amount *big.Int)
	// Adds a LOG to the state
	AddLog(*Log)
	// Adds a structured log to the env
	AddStructLog(StructLog)
	// Returns all coalesced structured logs
	StructLogs() []StructLog

	// Type of the VM
	VmType() Type

	// Current calling depth
	Depth() int
	SetDepth(i int)

	// Call another contract
	Call(me ContractRef, addr common.Address, data []byte, gas, price, value *big.Int) ([]byte, error)
	// Take another's contract code and execute within our own context
	CallCode(me ContractRef, addr common.Address, data []byte, gas, price, value *big.Int) ([]byte, error)
	// Create a new contract
	Create(me ContractRef, data []byte, gas, price, value *big.Int) ([]byte, common.Address, error)
}

// Database is a EVM database for full state querying
type Database interface {
	GetAccount(common.Address) Account
	CreateAccount(common.Address) Account

	AddBalance(common.Address, *big.Int)
	GetBalance(common.Address) *big.Int

	GetNonce(common.Address) uint64
	SetNonce(common.Address, uint64)

	GetCode(common.Address) []byte
	SetCode(common.Address, []byte)

	AddRefund(*big.Int)
	GetRefund() *big.Int

	GetState(common.Address, common.Hash) common.Hash
	SetState(common.Address, common.Hash, common.Hash)

	Delete(common.Address) bool
	Exist(common.Address) bool
	IsDeleted(common.Address) bool
}

// StructLog is emited to the Environment each cycle and lists information about the curent internal state
// prior to the execution of the statement.
type StructLog struct {
	Pc      uint64
	Op      OpCode
	Gas     *big.Int
	GasCost *big.Int
	Memory  []byte
	Stack   []*big.Int
	Storage map[common.Hash][]byte
	Err     error
}

type Account interface {
	SubBalance(amount *big.Int)
	AddBalance(amount *big.Int)
	SetBalance(*big.Int)
	SetNonce(uint64)
	Balance() *big.Int
	Address() common.Address
	ReturnGas(*big.Int, *big.Int)
	SetCode([]byte)
}
