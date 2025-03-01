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

package codec

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
	"strings"

	"github.com/cjminercn/go-coupe/rpc/shared"
)

const (
	READ_TIMEOUT      = 60 // in seconds
	MAX_REQUEST_SIZE  = 1024 * 1024
	MAX_RESPONSE_SIZE = 1024 * 1024
)

// Json serialization support
type JsonCodec struct {
	c net.Conn
	d *json.Decoder
}

// Create new JSON coder instance
func NewJsonCoder(conn net.Conn) ApiCoder {
	return &JsonCodec{
		c: conn,
		d: json.NewDecoder(conn),
	}
}

// Read incoming request and parse it to RPC request
func (self *JsonCodec) ReadRequest() (requests []*shared.Request, isBatch bool, err error) {
	deadline := time.Now().Add(READ_TIMEOUT * time.Second)
	if err := self.c.SetDeadline(deadline); err != nil {
		return nil, false, err
	}

	var incoming json.RawMessage
	err = self.d.Decode(&incoming)
	if err == nil {
		isBatch = incoming[0] == '['
		if isBatch {
			requests = make([]*shared.Request, 0)
			err = json.Unmarshal(incoming, &requests)
		} else {
			requests = make([]*shared.Request, 1)
			var singleRequest shared.Request
			if err = json.Unmarshal(incoming, &singleRequest); err == nil {
				requests[0] = &singleRequest
			}
		}
		return
	}

	self.c.Close()
	return nil, false, err
}

func (self *JsonCodec) Recv() (interface{}, error) {
	var msg json.RawMessage
	err := self.d.Decode(&msg)
	if err != nil {
		self.c.Close()
		return nil, err
	}

	return msg, err
}

func (self *JsonCodec) ReadResponse() (interface{}, error) {
	in, err := self.Recv()
	if err != nil {
		return nil, err
	}

	if msg, ok := in.(json.RawMessage); ok {
		var req *shared.Request
		if err = json.Unmarshal(msg, &req); err == nil && strings.HasPrefix(req.Method, "agent_") {
			return req, nil
		}

		var failure *shared.ErrorResponse
		if err = json.Unmarshal(msg, &failure); err == nil && failure.Error != nil {
			return failure, fmt.Errorf(failure.Error.Message)
		}

		var success *shared.SuccessResponse
		if err = json.Unmarshal(msg, &success); err == nil {
			return success, nil
		}
	}

	return in, err
}

// Decode data
func (self *JsonCodec) Decode(data []byte, msg interface{}) error {
	return json.Unmarshal(data, msg)
}

// Encode message
func (self *JsonCodec) Encode(msg interface{}) ([]byte, error) {
	return json.Marshal(msg)
}

// Parse JSON data from conn to obj
func (self *JsonCodec) WriteResponse(res interface{}) error {
	data, err := json.Marshal(res)
	if err != nil {
		self.c.Close()
		return err
	}

	bytesWritten := 0

	for bytesWritten < len(data) {
		n, err := self.c.Write(data[bytesWritten:])
		if err != nil {
			self.c.Close()
			return err
		}
		bytesWritten += n
	}

	return nil
}

// Close decoder and encoder
func (self *JsonCodec) Close() {
	self.c.Close()
}
