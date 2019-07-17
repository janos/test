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

import (
	"context"
	"sync"

	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/storage"
)

// BzzEth implements node.Service
var _ node.Service = &BzzEth{}

// BzzEth is a global module handling ethereum state on swarm
type BzzEth struct {
	mtx      sync.Mutex         // lock for peer pool
	peers    map[enode.ID]*Peer // peer pool
	netStore *storage.NetStore  // netstore to retrieve and store
	quit     chan struct{}      // quit channel to close go routines
}

// New constructs the BzzEth node service
func New(ns *storage.NetStore) *BzzEth {
	return &BzzEth{
		peers:    make(map[enode.ID]*Peer),
		netStore: ns,
		quit:     make(chan struct{}),
	}
}

func (b *BzzEth) getPeer(id enode.ID) *Peer {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	return b.peers[id]
}

func (b *BzzEth) addPeer(p *Peer) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	b.peers[p.ID()] = p
}

func (b *BzzEth) removePeer(p *Peer) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	delete(b.peers, p.ID())
}

// Run is the bzzeth protocol run function.
// - creates a peer
// - checks if it is a swarm node, put the protocol in idle mode
// - performs handshake
// - adds it the peerpool
// - starts handler loop
func (b *BzzEth) Run(p *p2p.Peer, rw p2p.MsgReadWriter) error {
	peer := protocols.NewPeer(p, rw, Spec)
	bp := NewPeer(peer)

	// perform handshake and register if peer serves headers
	handshake, err := bp.Handshake(context.TODO(), Handshake{ServeHeaders: true}, nil)
	if err != nil {
		return err
	}
	bp.serveHeaders = handshake.(*Handshake).ServeHeaders
	log.Warn("handshake", "hs", handshake, "peer", bp)
	// with another swarm node the protocol goes into idle
	if isSwarmNodeFunc(bp) {
		<-b.quit
		return nil
	}
	b.addPeer(bp)
	defer b.removePeer(bp)

	return peer.Run(b.HandleMsg(bp))
}

// HandleMsg is the message handler that delegates incoming messages
func (b *BzzEth) HandleMsg(p *Peer) func(context.Context, interface{}) error {
	return func(ctx context.Context, msg interface{}) error {
		log.Info("handling msg", "peer", p.ID())
		switch msg := msg.(type) {
		case *NewBlockHeaders:
			go b.handleNewBlockHeaders(ctx, p, msg)
		case *GetBlockHeaders:
			go b.handleGetBlockHeaders(ctx, p, msg)
		case *BlockHeaders:
			go b.handleBlockHeaders(ctx, p, msg)
		}
		return nil
	}
}

func (b *BzzEth) handleNewBlockHeaders(ctx context.Context, p *Peer, msg *NewBlockHeaders) {
	log.Debug("bzzeth.handleNewBlockHeaders")
}

func (b *BzzEth) handleGetBlockHeaders(ctx context.Context, p *Peer, msg *GetBlockHeaders) {
	log.Debug("bzzeth.handleGetBlockHeaders")
}

func (b *BzzEth) handleBlockHeaders(ctx context.Context, p *Peer, msg *BlockHeaders) {
	log.Debug("bzzeth.handleBlockHeaders")
}

// Protocols returns the p2p protocol
func (b *BzzEth) Protocols() []p2p.Protocol {
	return []p2p.Protocol{
		{
			Name:    Spec.Name,
			Version: Spec.Version,
			Length:  Spec.Length(),
			Run:     b.Run,
		},
	}
}

// APIs return APIs defined on the node service
func (b *BzzEth) APIs() []rpc.API {
	return nil
}

// Start starts the BzzEth node service
func (b *BzzEth) Start(server *p2p.Server) error {
	log.Info("bzzeth starting...")
	return nil
}

// Stop stops the BzzEth node service
func (b *BzzEth) Stop() error {
	log.Info("bzzeth shutting down...")
	close(b.quit)
	return nil
}

// this function is called to check if the remote peer is another swarm node
// in which case the protocol is idle
// can be reassigned in test to mock a swarm node
var isSwarmNodeFunc = isSwarmNode

func isSwarmNode(p *Peer) bool {
	return p.HasCap("bzz")
}
