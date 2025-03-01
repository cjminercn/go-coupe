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

package api

import (
	"fmt"
	"time"

	"github.com/cjminercn/go-coupe/common"
	"github.com/cjminercn/go-coupe/eth"
	"github.com/cjminercn/go-coupe/rpc/codec"
	"github.com/cjminercn/go-coupe/rpc/shared"
	"github.com/cjminercn/go-coupe/xeth"
)

const (
	PersonalApiVersion = "1.0"
)

var (
	// mapping between methods and handlers
	personalMapping = map[string]personalhandler{
		"personal_listAccounts":  (*personalApi).ListAccounts,
		"personal_newAccount":    (*personalApi).NewAccount,
		"personal_unlockAccount": (*personalApi).UnlockAccount,
	}
)

// net callback handler
type personalhandler func(*personalApi, *shared.Request) (interface{}, error)

// net api provider
type personalApi struct {
	xeth     *xeth.XEth
	cjminercn *eth.cjminercn
	methods  map[string]personalhandler
	codec    codec.ApiCoder
}

// create a new net api instance
func NewPersonalApi(xeth *xeth.XEth, eth *eth.cjminercn, coder codec.Codec) *personalApi {
	return &personalApi{
		xeth:     xeth,
		cjminercn: eth,
		methods:  personalMapping,
		codec:    coder.New(nil),
	}
}

// collection with supported methods
func (self *personalApi) Methods() []string {
	methods := make([]string, len(self.methods))
	i := 0
	for k := range self.methods {
		methods[i] = k
		i++
	}
	return methods
}

// Execute given request
func (self *personalApi) Execute(req *shared.Request) (interface{}, error) {
	if callback, ok := self.methods[req.Method]; ok {
		return callback(self, req)
	}

	return nil, shared.NewNotImplementedError(req.Method)
}

func (self *personalApi) Name() string {
	return shared.PersonalApiName
}

func (self *personalApi) ApiVersion() string {
	return PersonalApiVersion
}

func (self *personalApi) ListAccounts(req *shared.Request) (interface{}, error) {
	return self.xeth.Accounts(), nil
}

func (self *personalApi) NewAccount(req *shared.Request) (interface{}, error) {
	args := new(NewAccountArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}
	var passwd string
	if args.Passphrase == nil {
		fe := self.xeth.Frontend()
		if fe == nil {
			return false, fmt.Errorf("unable to create account: unable to interact with user")
		}
		var ok bool
		passwd, ok = fe.AskPassword()
		if !ok {
			return false, fmt.Errorf("unable to create account: no password given")
		}
	} else {
		passwd = *args.Passphrase
	}
	am := self.cjminercn.AccountManager()
	acc, err := am.NewAccount(passwd)
	return acc.Address.Hex(), err
}

func (self *personalApi) UnlockAccount(req *shared.Request) (interface{}, error) {
	args := new(UnlockAccountArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	if args.Passphrase == nil {
		fe := self.xeth.Frontend()
		if fe == nil {
			return false, fmt.Errorf("No password provided")
		}
		return fe.UnlockAccount(common.HexToAddress(args.Address).Bytes()), nil
	}

	am := self.cjminercn.AccountManager()
	addr := common.HexToAddress(args.Address)

	err := am.TimedUnlock(addr, *args.Passphrase, time.Duration(args.Duration)*time.Second)
	return err == nil, err
}
