// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package executor

import (
	"github.com/dim4egster/avalanchego/ids"
	"github.com/dim4egster/avalanchego/snow"
	"github.com/dim4egster/avalanchego/utils"
	"github.com/dim4egster/avalanchego/vms/platformvm/blocks"
	"github.com/dim4egster/avalanchego/vms/platformvm/state"
	"github.com/dim4egster/avalanchego/vms/platformvm/txs/mempool"
)

// Shared fields used by visitors.
type backend struct {
	mempool.Mempool
	// Keep the last accepted block in memory because when we check a
	// proposal block's status, it may be accepted but not have an accepted
	// child, in which case it's in [blkIDToState].
	lastAccepted ids.ID

	// blkIDToState is a map from a block's ID to the state of the block.
	// Blocks are put into this map when they are verified.
	// Proposal blocks are removed from this map when they are rejected
	// or when a child is accepted.
	// All other blocks are removed when they are accepted/rejected.
	// Note that Genesis block is a commit block so no need to update
	// blkIDToState with it upon backend creation (Genesis is already accepted)
	blkIDToState map[ids.ID]*blockState
	state        state.State

	ctx          *snow.Context
	bootstrapped *utils.AtomicBool
}

func (b *backend) GetState(blkID ids.ID) (state.Chain, bool) {
	// If the block is in the map, it is either processing or a proposal block
	// that was accepted without an accepted child.
	if state, ok := b.blkIDToState[blkID]; ok {
		if state.onAcceptState != nil {
			return state.onAcceptState, true
		}
		return nil, false
	}

	// Note: If the last accepted block is a proposal block, we will have
	//       returned in the above if statement.
	return b.state, blkID == b.lastAccepted
}

func (b *backend) GetBlock(blkID ids.ID) (blocks.Block, error) {
	// See if the block is in memory.
	if blk, ok := b.blkIDToState[blkID]; ok {
		return blk.statelessBlock, nil
	}
	// The block isn't in memory. Check the database.
	statelessBlk, _, err := b.state.GetStatelessBlock(blkID)
	return statelessBlk, err
}

func (b *backend) LastAccepted() ids.ID {
	return b.lastAccepted
}

func (b *backend) free(blkID ids.ID) {
	delete(b.blkIDToState, blkID)
}
