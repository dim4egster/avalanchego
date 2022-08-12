// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package blocks

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dim4egster/avalanchego/codec"
	"github.com/dim4egster/avalanchego/ids"
	"github.com/dim4egster/avalanchego/utils/crypto"
	"github.com/dim4egster/avalanchego/vms/components/avax"
	"github.com/dim4egster/avalanchego/vms/platformvm/txs"
	"github.com/dim4egster/avalanchego/vms/secp256k1fx"
)

var preFundedKeys = crypto.BuildTestKeys()

func TestStandardBlock(t *testing.T) {
	// check standard block can be built and parsed
	assert := assert.New(t)
	parentID := ids.ID{'p', 'a', 'r', 'e', 'n', 't', 'I', 'D'}
	height := uint64(2022)
	txs, err := testDecisionTxs()
	assert.NoError(err)

	for _, cdc := range []codec.Manager{Codec, GenesisCodec} {
		// build block
		standardBlk, err := NewStandardBlock(
			parentID,
			height,
			txs,
		)
		assert.NoError(err)

		// parse block
		parsed, err := Parse(cdc, standardBlk.Bytes())
		assert.NoError(err)

		// compare content
		assert.Equal(standardBlk.ID(), parsed.ID())
		assert.Equal(standardBlk.Bytes(), parsed.Bytes())
		assert.Equal(standardBlk.Parent(), parsed.Parent())
		assert.Equal(standardBlk.Height(), parsed.Height())

		parsedStandardBlk, ok := parsed.(*StandardBlock)
		assert.True(ok)
		assert.Equal(txs, parsedStandardBlk.Transactions)
	}
}

func TestProposalBlock(t *testing.T) {
	// check proposal block can be built and parsed
	assert := assert.New(t)
	parentID := ids.ID{'p', 'a', 'r', 'e', 'n', 't', 'I', 'D'}
	height := uint64(2022)
	tx, err := testProposalTx()
	assert.NoError(err)

	for _, cdc := range []codec.Manager{Codec, GenesisCodec} {
		// build block
		proposalBlk, err := NewProposalBlock(
			parentID,
			height,
			tx,
		)
		assert.NoError(err)

		// parse block
		parsed, err := Parse(cdc, proposalBlk.Bytes())
		assert.NoError(err)

		// compare content
		assert.Equal(proposalBlk.ID(), parsed.ID())
		assert.Equal(proposalBlk.Bytes(), parsed.Bytes())
		assert.Equal(proposalBlk.Parent(), parsed.Parent())
		assert.Equal(proposalBlk.Height(), parsed.Height())

		parsedProposalBlk, ok := parsed.(*ProposalBlock)
		assert.True(ok)
		assert.Equal(tx, parsedProposalBlk.Tx)
	}
}

func TestCommitBlock(t *testing.T) {
	// check commit block can be built and parsed
	assert := assert.New(t)
	parentID := ids.ID{'p', 'a', 'r', 'e', 'n', 't', 'I', 'D'}
	height := uint64(2022)

	for _, cdc := range []codec.Manager{Codec, GenesisCodec} {
		// build block
		commitBlk, err := NewCommitBlock(parentID, height)
		assert.NoError(err)

		// parse block
		parsed, err := Parse(cdc, commitBlk.Bytes())
		assert.NoError(err)

		// compare content
		assert.Equal(commitBlk.ID(), parsed.ID())
		assert.Equal(commitBlk.Bytes(), parsed.Bytes())
		assert.Equal(commitBlk.Parent(), parsed.Parent())
		assert.Equal(commitBlk.Height(), parsed.Height())
	}
}

func TestAbortBlock(t *testing.T) {
	// check abort block can be built and parsed
	assert := assert.New(t)
	parentID := ids.ID{'p', 'a', 'r', 'e', 'n', 't', 'I', 'D'}
	height := uint64(2022)

	for _, cdc := range []codec.Manager{Codec, GenesisCodec} {
		// build block
		abortBlk, err := NewAbortBlock(parentID, height)
		assert.NoError(err)

		// parse block
		parsed, err := Parse(cdc, abortBlk.Bytes())
		assert.NoError(err)

		// compare content
		assert.Equal(abortBlk.ID(), parsed.ID())
		assert.Equal(abortBlk.Bytes(), parsed.Bytes())
		assert.Equal(abortBlk.Parent(), parsed.Parent())
		assert.Equal(abortBlk.Height(), parsed.Height())
	}
}

func TestAtomicBlock(t *testing.T) {
	// check atomic block can be built and parsed
	assert := assert.New(t)
	parentID := ids.ID{'p', 'a', 'r', 'e', 'n', 't', 'I', 'D'}
	height := uint64(2022)
	tx, err := testAtomicTx()
	assert.NoError(err)

	for _, cdc := range []codec.Manager{Codec, GenesisCodec} {
		// build block
		atomicBlk, err := NewAtomicBlock(
			parentID,
			height,
			tx,
		)
		assert.NoError(err)

		// parse block
		parsed, err := Parse(cdc, atomicBlk.Bytes())
		assert.NoError(err)

		// compare content
		assert.Equal(atomicBlk.ID(), parsed.ID())
		assert.Equal(atomicBlk.Bytes(), parsed.Bytes())
		assert.Equal(atomicBlk.Parent(), parsed.Parent())
		assert.Equal(atomicBlk.Height(), parsed.Height())

		parsedAtomicBlk, ok := parsed.(*AtomicBlock)
		assert.True(ok)
		assert.Equal(tx, parsedAtomicBlk.Tx)
	}
}

func testAtomicTx() (*txs.Tx, error) {
	utx := &txs.ImportTx{
		BaseTx: txs.BaseTx{BaseTx: avax.BaseTx{
			NetworkID:    10,
			BlockchainID: ids.ID{'c', 'h', 'a', 'i', 'n', 'I', 'D'},
			Outs: []*avax.TransferableOutput{{
				Asset: avax.Asset{ID: ids.ID{'a', 's', 's', 'e', 'r', 't'}},
				Out: &secp256k1fx.TransferOutput{
					Amt: uint64(1234),
					OutputOwners: secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     []ids.ShortID{preFundedKeys[0].PublicKey().Address()},
					},
				},
			}},
			Ins: []*avax.TransferableInput{{
				UTXOID: avax.UTXOID{
					TxID:        ids.ID{'t', 'x', 'I', 'D'},
					OutputIndex: 2,
				},
				Asset: avax.Asset{ID: ids.ID{'a', 's', 's', 'e', 'r', 't'}},
				In: &secp256k1fx.TransferInput{
					Amt:   uint64(5678),
					Input: secp256k1fx.Input{SigIndices: []uint32{0}},
				},
			}},
			Memo: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		}},
		SourceChain: ids.ID{'c', 'h', 'a', 'i', 'n'},
		ImportedInputs: []*avax.TransferableInput{{
			UTXOID: avax.UTXOID{
				TxID:        ids.Empty.Prefix(1),
				OutputIndex: 1,
			},
			Asset: avax.Asset{ID: ids.ID{'a', 's', 's', 'e', 'r', 't'}},
			In: &secp256k1fx.TransferInput{
				Amt:   50000,
				Input: secp256k1fx.Input{SigIndices: []uint32{0}},
			},
		}},
	}
	signers := [][]*crypto.PrivateKeySECP256K1R{{preFundedKeys[0]}}
	return txs.NewSigned(utx, txs.Codec, signers)
}

func testDecisionTxs() ([]*txs.Tx, error) {
	countTxs := 2
	txes := make([]*txs.Tx, 0, countTxs)
	for i := 0; i < countTxs; i++ {
		// Create the tx
		utx := &txs.CreateChainTx{
			BaseTx: txs.BaseTx{BaseTx: avax.BaseTx{
				NetworkID:    10,
				BlockchainID: ids.ID{'c', 'h', 'a', 'i', 'n', 'I', 'D'},
				Outs: []*avax.TransferableOutput{{
					Asset: avax.Asset{ID: ids.ID{'a', 's', 's', 'e', 'r', 't'}},
					Out: &secp256k1fx.TransferOutput{
						Amt: uint64(1234),
						OutputOwners: secp256k1fx.OutputOwners{
							Threshold: 1,
							Addrs:     []ids.ShortID{preFundedKeys[0].PublicKey().Address()},
						},
					},
				}},
				Ins: []*avax.TransferableInput{{
					UTXOID: avax.UTXOID{
						TxID:        ids.ID{'t', 'x', 'I', 'D'},
						OutputIndex: 2,
					},
					Asset: avax.Asset{ID: ids.ID{'a', 's', 's', 'e', 'r', 't'}},
					In: &secp256k1fx.TransferInput{
						Amt:   uint64(5678),
						Input: secp256k1fx.Input{SigIndices: []uint32{0}},
					},
				}},
				Memo: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			}},
			SubnetID:    ids.ID{'s', 'u', 'b', 'n', 'e', 't', 'I', 'D'},
			ChainName:   "a chain",
			VMID:        ids.GenerateTestID(),
			FxIDs:       []ids.ID{ids.GenerateTestID()},
			GenesisData: []byte{'g', 'e', 'n', 'D', 'a', 't', 'a'},
			SubnetAuth:  &secp256k1fx.Input{SigIndices: []uint32{1}},
		}

		signers := [][]*crypto.PrivateKeySECP256K1R{{preFundedKeys[0]}}
		tx, err := txs.NewSigned(utx, txs.Codec, signers)
		if err != nil {
			return nil, err
		}
		txes = append(txes, tx)
	}
	return txes, nil
}

func testProposalTx() (*txs.Tx, error) {
	utx := &txs.RewardValidatorTx{
		TxID: ids.ID{'r', 'e', 'w', 'a', 'r', 'd', 'I', 'D'},
	}

	signers := [][]*crypto.PrivateKeySECP256K1R{{preFundedKeys[0]}}
	return txs.NewSigned(utx, txs.Codec, signers)
}
