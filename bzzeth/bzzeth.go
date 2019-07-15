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

var Spec = &protocols.Spec{
	Name:       "bzzeth",
	Version:    1,
	MaxMsgSize: 10 * 1024 * 1024,
	Messages: []interface{}{
		Handshake{},
		NewBlockHeaders{},
		GetBlockHeaders{},
		BlockHeaders{},
	},
}

type BzzEth struct {
	mtx   sync.Mutex
	peers map[enode.ID]*Peer
	spec  *protocols.Spec

	netStore *storage.NetStore

	quit chan struct{}
}

func New(ns *storage.NetStore) *BzzEth {
	bzzeth := &BzzEth{
		peers:    make(map[enode.ID]*Peer),
		netStore: ns,
		quit:     make(chan struct{}),
	}

	bzzeth.spec = Spec

	return bzzeth
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

func (b *BzzEth) Run(p *p2p.Peer, rw p2p.MsgReadWriter) error {
	peer := protocols.NewPeer(p, rw, b.spec)
	bp := NewPeer(peer)
	b.addPeer(bp)

	defer b.removePeer(bp)
	defer close(bp.quit)

	handshake, err := bp.Handshake(context.TODO(), Handshake{ServeHeaders: true}, nil)
	if err != nil {
		return err
	}
	bp.servesHeaders = handshake.(*Handshake).ServeHeaders
	if isSwarmNodeFunc(bp) {
		// swarm node - do nothing
		<-b.quit
		return nil
	}

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

func (b *BzzEth) Protocols() []p2p.Protocol {
	return []p2p.Protocol{
		{
			Name:    "bzzeth",
			Version: 1,
			Length:  10 * 1024 * 1024,
			Run:     b.Run,
		},
	}
}

func (b *BzzEth) APIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: "bzzeth",
			Version:   "1.0",
			Service:   NewAPI(b),
			Public:    false,
		},
	}
}

// Additional public methods accessible through API for pss
type API struct {
	*BzzEth
}

func NewAPI(b *BzzEth) *API {
	return &API{BzzEth: b}
}

func (b *BzzEth) Start(server *p2p.Server) error {
	log.Info("bzz-eth starting...")
	return nil
}

func (b *BzzEth) Stop() error {
	log.Info("bzz-eth shutting down...")
	return nil
}

var isSwarmNodeFunc = isSwarmNode

func isSwarmNode(p *Peer) (yes bool) {
	return p.HasCap("bzz")
}
