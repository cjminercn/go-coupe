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

package comms

import (
	"fmt"

	"github.com/cjminercn/go-coupe/rpc/codec"
	"github.com/cjminercn/go-coupe/rpc/shared"
)

type InProcClient struct {
	api         shared.cjminercnApi
	codec       codec.Codec
	lastId      interface{}
	lastJsonrpc string
	lastErr     error
	lastRes     interface{}
}

// Create a new in process client
func NewInProcClient(codec codec.Codec) *InProcClient {
	return &InProcClient{
		codec: codec,
	}
}

func (self *InProcClient) Close() {
	// do nothing
}

// Need to setup api support
func (self *InProcClient) Initialize(offeredApi shared.cjminercnApi) {
	self.api = offeredApi
}

func (self *InProcClient) Send(req interface{}) error {
	if r, ok := req.(*shared.Request); ok {
		self.lastId = r.Id
		self.lastJsonrpc = r.Jsonrpc
		self.lastRes, self.lastErr = self.api.Execute(r)
		return self.lastErr
	}

	return fmt.Errorf("Invalid request (%T)", req)
}

func (self *InProcClient) Recv() (interface{}, error) {
	return *shared.NewRpcResponse(self.lastId, self.lastJsonrpc, self.lastRes, self.lastErr), nil
}

func (self *InProcClient) SupportedModules() (map[string]string, error) {
	req := shared.Request{
		Id:      1,
		Jsonrpc: "2.0",
		Method:  "modules",
	}

	if res, err := self.api.Execute(&req); err == nil {
		if result, ok := res.(map[string]string); ok {
			return result, nil
		}
	} else {
		return nil, err
	}

	return nil, fmt.Errorf("Invalid response")
}
