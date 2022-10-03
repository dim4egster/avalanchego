// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package blocks

import (
	"time"

	"github.com/dim4egster/qmallgo/ids"
	"github.com/dim4egster/qmallgo/snow"
	"github.com/dim4egster/qmallgo/vms/platformvm/txs"
)

var (
	_ BlueberryBlock = &BlueberryCommitBlock{}
	_ Block          = &ApricotCommitBlock{}
)

type BlueberryCommitBlock struct {
	Time               uint64 `serialize:"true" json:"time"`
	ApricotCommitBlock `serialize:"true"`
}

func (b *BlueberryCommitBlock) Timestamp() time.Time  { return time.Unix(int64(b.Time), 0) }
func (b *BlueberryCommitBlock) Visit(v Visitor) error { return v.BlueberryCommitBlock(b) }

func NewBlueberryCommitBlock(
	timestamp time.Time,
	parentID ids.ID,
	height uint64,
) (*BlueberryCommitBlock, error) {
	blk := &BlueberryCommitBlock{
		Time: uint64(timestamp.Unix()),
		ApricotCommitBlock: ApricotCommitBlock{
			CommonBlock: CommonBlock{
				PrntID: parentID,
				Hght:   height,
			},
		},
	}
	return blk, initialize(blk)
}

type ApricotCommitBlock struct {
	CommonBlock `serialize:"true"`
}

func (b *ApricotCommitBlock) initialize(bytes []byte) error {
	b.CommonBlock.initialize(bytes)
	return nil
}

func (*ApricotCommitBlock) InitCtx(ctx *snow.Context) {}

func (*ApricotCommitBlock) Txs() []*txs.Tx          { return nil }
func (b *ApricotCommitBlock) Visit(v Visitor) error { return v.ApricotCommitBlock(b) }

func NewApricotCommitBlock(
	parentID ids.ID,
	height uint64,
) (*ApricotCommitBlock, error) {
	blk := &ApricotCommitBlock{
		CommonBlock: CommonBlock{
			PrntID: parentID,
			Hght:   height,
		},
	}
	return blk, initialize(blk)
}
