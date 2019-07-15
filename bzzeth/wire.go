// Copyright 2019 The Swarm Authors
// This file is part of the Swarm library.
//
// The Swarm library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Swarm library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Swarm library. If not, see <http://www.gnu.org/licenses/>.

package bzzeth

// Handshake is used in between the ethereum node and the Swarm node
type Handshake struct {
	ServeHeaders bool //indicates if this node is expected to serve requests for headers
}

// NewBlockHeaders is sent from the Ethereum client to the Swarm node
type NewBlockHeaders struct {
	Headers []Header
}

type Header struct {
	Hash   []byte //block hash
	Number []byte //block height
}

// GetBlockHeader is used between a Swarm node and the Ethereum node in two cases:
// 1. When an Ethereum node asks the header corresponding to the hashes in the message(eth -> bzz)
// 2. When a Swarm node cannot find a particular header in the network, it asks the ethereum node for the header in order to push it to the network (bzz -> eth)
type GetBlockHeaders struct {
	Id     uint32
	Hashes [][]byte
}

type BlockHeaders struct {
	Id      uint32
	Headers [][]byte
}
