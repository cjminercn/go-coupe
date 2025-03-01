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
	"net"
	"strconv"

	"github.com/cjminercn/go-coupe/rpc/shared"
)

type Codec int

// (de)serialization support for rpc interface
type ApiCoder interface {
	// Parse message to request from underlying stream
	ReadRequest() ([]*shared.Request, bool, error)
	// Parse response message from underlying stream
	ReadResponse() (interface{}, error)
	// Read raw message from underlying stream
	Recv() (interface{}, error)
	// Encode response to encoded form in underlying stream
	WriteResponse(interface{}) error
	// Decode single message from data
	Decode([]byte, interface{}) error
	// Encode msg to encoded form
	Encode(msg interface{}) ([]byte, error)
	// close the underlying stream
	Close()
}

// supported codecs
const (
	JSON Codec = iota
	nCodecs
)

var (
	// collection with supported coders
	coders = make([]func(net.Conn) ApiCoder, nCodecs)
)

// create a new coder instance
func (c Codec) New(conn net.Conn) ApiCoder {
	switch c {
	case JSON:
		return NewJsonCoder(conn)
	}

	panic("codec: request for codec #" + strconv.Itoa(int(c)) + " is unavailable")
}
