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
	"github.com/cjminercn/go-coupe/eth"
	"github.com/cjminercn/go-coupe/rpc/codec"
	"github.com/cjminercn/go-coupe/rpc/shared"
	"github.com/cjminercn/go-coupe/xeth"
)

const (
	DbApiversion = "1.0"
)

var (
	// mapping between methods and handlers
	DbMapping = map[string]dbhandler{
		"db_getString": (*dbApi).GetString,
		"db_putString": (*dbApi).PutString,
		"db_getHex":    (*dbApi).GetHex,
		"db_putHex":    (*dbApi).PutHex,
	}
)

// db callback handler
type dbhandler func(*dbApi, *shared.Request) (interface{}, error)

// db api provider
type dbApi struct {
	xeth     *xeth.XEth
	cjminercn *eth.cjminercn
	methods  map[string]dbhandler
	codec    codec.ApiCoder
}

// create a new db api instance
func NewDbApi(xeth *xeth.XEth, cjminercn *eth.cjminercn, coder codec.Codec) *dbApi {
	return &dbApi{
		xeth:     xeth,
		cjminercn: cjminercn,
		methods:  DbMapping,
		codec:    coder.New(nil),
	}
}

// collection with supported methods
func (self *dbApi) Methods() []string {
	methods := make([]string, len(self.methods))
	i := 0
	for k := range self.methods {
		methods[i] = k
		i++
	}
	return methods
}

// Execute given request
func (self *dbApi) Execute(req *shared.Request) (interface{}, error) {
	if callback, ok := self.methods[req.Method]; ok {
		return callback(self, req)
	}

	return nil, &shared.NotImplementedError{req.Method}
}

func (self *dbApi) Name() string {
	return shared.DbApiName
}

func (self *dbApi) ApiVersion() string {
	return DbApiversion
}

func (self *dbApi) GetString(req *shared.Request) (interface{}, error) {
	args := new(DbArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	if err := args.requirements(); err != nil {
		return nil, err
	}

	ret, err := self.xeth.DbGet([]byte(args.Database + args.Key))
	return string(ret), err
}

func (self *dbApi) PutString(req *shared.Request) (interface{}, error) {
	args := new(DbArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	if err := args.requirements(); err != nil {
		return nil, err
	}

	return self.xeth.DbPut([]byte(args.Database+args.Key), args.Value), nil
}

func (self *dbApi) GetHex(req *shared.Request) (interface{}, error) {
	args := new(DbHexArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	if err := args.requirements(); err != nil {
		return nil, err
	}

	if res, err := self.xeth.DbGet([]byte(args.Database + args.Key)); err == nil {
		return newHexData(res), nil
	} else {
		return nil, err
	}
}

func (self *dbApi) PutHex(req *shared.Request) (interface{}, error) {
	args := new(DbHexArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	if err := args.requirements(); err != nil {
		return nil, err
	}

	return self.xeth.DbPut([]byte(args.Database+args.Key), args.Value), nil
}
