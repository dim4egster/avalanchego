// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package executor

import (
	"fmt"

	"github.com/dim4egster/qmallgo/snow/consensus/snowman"
	"github.com/dim4egster/qmallgo/vms/platformvm/blocks"
)

var _ blocks.Visitor = &verifier{}

// options supports build new option blocks
type options struct {
	// outputs populated by this struct's methods:
	commitBlock blocks.Block
	abortBlock  blocks.Block
}

func (*options) BlueberryAbortBlock(*blocks.BlueberryAbortBlock) error {
	return snowman.ErrNotOracle
}

func (*options) BlueberryCommitBlock(*blocks.BlueberryCommitBlock) error {
	return snowman.ErrNotOracle
}

func (o *options) BlueberryProposalBlock(b *blocks.BlueberryProposalBlock) error {
	timestamp := b.Timestamp()
	blkID := b.ID()
	nextHeight := b.Height() + 1

	var err error
	o.commitBlock, err = blocks.NewBlueberryCommitBlock(timestamp, blkID, nextHeight)
	if err != nil {
		return fmt.Errorf(
			"failed to create commit block: %w",
			err,
		)
	}

	o.abortBlock, err = blocks.NewBlueberryAbortBlock(timestamp, blkID, nextHeight)
	if err != nil {
		return fmt.Errorf(
			"failed to create abort block: %w",
			err,
		)
	}
	return nil
}

func (*options) BlueberryStandardBlock(*blocks.BlueberryStandardBlock) error {
	return snowman.ErrNotOracle
}

func (*options) ApricotAbortBlock(*blocks.ApricotAbortBlock) error {
	return snowman.ErrNotOracle
}

func (*options) ApricotCommitBlock(*blocks.ApricotCommitBlock) error {
	return snowman.ErrNotOracle
}

func (o *options) ApricotProposalBlock(b *blocks.ApricotProposalBlock) error {
	blkID := b.ID()
	nextHeight := b.Height() + 1

	var err error
	o.commitBlock, err = blocks.NewApricotCommitBlock(blkID, nextHeight)
	if err != nil {
		return fmt.Errorf(
			"failed to create commit block: %w",
			err,
		)
	}

	o.abortBlock, err = blocks.NewApricotAbortBlock(blkID, nextHeight)
	if err != nil {
		return fmt.Errorf(
			"failed to create abort block: %w",
			err,
		)
	}
	return nil
}

func (*options) ApricotStandardBlock(*blocks.ApricotStandardBlock) error {
	return snowman.ErrNotOracle
}

func (*options) ApricotAtomicBlock(*blocks.ApricotAtomicBlock) error {
	return snowman.ErrNotOracle
}
