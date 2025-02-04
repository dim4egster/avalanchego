// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package executor

import (
	"go.uber.org/zap"

	"github.com/dim4egster/qmallgo/snow/choices"
	"github.com/dim4egster/qmallgo/vms/platformvm/blocks"
)

var _ blocks.Visitor = &rejector{}

// rejector handles the logic for rejecting a block.
// All errors returned by this struct are fatal and should result in the chain
// being shutdown.
type rejector struct {
	*backend
}

func (r *rejector) BlueberryAbortBlock(b *blocks.BlueberryAbortBlock) error {
	return r.rejectBlock(b, "blueberry abort")
}

func (r *rejector) BlueberryCommitBlock(b *blocks.BlueberryCommitBlock) error {
	return r.rejectBlock(b, "blueberry commit")
}

func (r *rejector) BlueberryProposalBlock(b *blocks.BlueberryProposalBlock) error {
	return r.rejectBlock(b, "blueberry proposal")
}

func (r *rejector) BlueberryStandardBlock(b *blocks.BlueberryStandardBlock) error {
	return r.rejectBlock(b, "blueberry standard")
}

func (r *rejector) ApricotAbortBlock(b *blocks.ApricotAbortBlock) error {
	return r.rejectBlock(b, "apricot abort")
}

func (r *rejector) ApricotCommitBlock(b *blocks.ApricotCommitBlock) error {
	return r.rejectBlock(b, "apricot commit")
}

func (r *rejector) ApricotProposalBlock(b *blocks.ApricotProposalBlock) error {
	return r.rejectBlock(b, "apricot proposal")
}

func (r *rejector) ApricotStandardBlock(b *blocks.ApricotStandardBlock) error {
	return r.rejectBlock(b, "apricot standard")
}

func (r *rejector) ApricotAtomicBlock(b *blocks.ApricotAtomicBlock) error {
	return r.rejectBlock(b, "apricot atomic")
}

func (r *rejector) rejectBlock(b blocks.Block, blockType string) error {
	blkID := b.ID()
	defer r.free(blkID)

	r.ctx.Log.Verbo(
		"rejecting block",
		zap.String("blockType", blockType),
		zap.Stringer("blkID", blkID),
		zap.Uint64("height", b.Height()),
		zap.Stringer("parentID", b.Parent()),
	)

	for _, tx := range b.Txs() {
		if err := r.Mempool.Add(tx); err != nil {
			r.ctx.Log.Debug(
				"failed to reissue tx",
				zap.Stringer("txID", tx.ID()),
				zap.Stringer("blkID", blkID),
				zap.Error(err),
			)
		}
	}

	r.state.AddStatelessBlock(b, choices.Rejected)
	return r.state.Commit()
}
