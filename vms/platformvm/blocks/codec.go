// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package blocks

import (
	"math"

	"github.com/dim4egster/qmallgo/codec"
	"github.com/dim4egster/qmallgo/codec/linearcodec"
	"github.com/dim4egster/qmallgo/utils/wrappers"
	"github.com/dim4egster/qmallgo/vms/platformvm/txs"
)

// GenesisCode allows blocks of larger than usual size to be parsed.
// While this gives flexibility in accommodating large genesis blocks
// it must not be used to parse new, unverified blocks which instead
// must be processed by Codec
var (
	Codec        codec.Manager
	GenesisCodec codec.Manager
)

func init() {
	c := linearcodec.NewDefault()
	Codec = codec.NewDefaultManager()
	gc := linearcodec.NewCustomMaxLength(math.MaxInt32)
	GenesisCodec = codec.NewManager(math.MaxInt32)

	errs := wrappers.Errs{}
	for _, c := range []codec.Registry{c, gc} {
		errs.Add(
			RegisterApricotBlockTypes(c),
			txs.RegisterUnsignedTxsTypes(c),
			RegisterBlueberryBlockTypes(c),
		)
	}
	errs.Add(
		Codec.RegisterCodec(txs.Version, c),
		GenesisCodec.RegisterCodec(txs.Version, gc),
	)
	if errs.Errored() {
		panic(errs.Err)
	}
}

// RegisterApricotBlockTypes allows registering relevant type of blocks package
// in the right sequence. Following repackaging of platformvm package, a few
// subpackage-level codecs were introduced, each handling serialization of
// specific types.
func RegisterApricotBlockTypes(targetCodec codec.Registry) error {
	errs := wrappers.Errs{}
	errs.Add(
		targetCodec.RegisterType(&ApricotProposalBlock{}),
		targetCodec.RegisterType(&ApricotAbortBlock{}),
		targetCodec.RegisterType(&ApricotCommitBlock{}),
		targetCodec.RegisterType(&ApricotStandardBlock{}),
		targetCodec.RegisterType(&ApricotAtomicBlock{}),
	)
	return errs.Err
}

func RegisterBlueberryBlockTypes(targetCodec codec.Registry) error {
	errs := wrappers.Errs{}
	errs.Add(
		targetCodec.RegisterType(&BlueberryProposalBlock{}),
		targetCodec.RegisterType(&BlueberryAbortBlock{}),
		targetCodec.RegisterType(&BlueberryCommitBlock{}),
		targetCodec.RegisterType(&BlueberryStandardBlock{}),
	)
	return errs.Err
}
