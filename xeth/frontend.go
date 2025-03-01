// Copyright 2015 The go-coupe Authors
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

package xeth

// Frontend should be implemented by users of XEth. Its methods are
// called whenever XEth makes a decision that requires user input.
type Frontend interface {
	// AskPassword is called when a new account is created or updated
	AskPassword() (string, bool)

	// UnlockAccount is called when a transaction needs to be signed
	// but the key corresponding to the transaction's sender is
	// locked.
	//
	// It should unlock the account with the given address and return
	// true if unlocking succeeded.
	UnlockAccount(address []byte) bool

	// This is called for all transactions inititated through
	// Transact. It should prompt the user to confirm the transaction
	// and return true if the transaction was acknowledged.
	//
	// ConfirmTransaction is not used for Call transactions
	// because they cannot change any state.
	ConfirmTransaction(tx string) bool
}

// dummyFrontend is a non-interactive frontend that allows all
// transactions but cannot not unlock any keys.
type dummyFrontend struct{}

func (dummyFrontend) AskPassword() (string, bool)    { return "", false }
func (dummyFrontend) UnlockAccount([]byte) bool      { return false }
func (dummyFrontend) ConfirmTransaction(string) bool { return true }
