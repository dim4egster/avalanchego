// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package states

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/stretchr/testify/require"

	"github.com/dim4egster/qmallgo/database"
	"github.com/dim4egster/qmallgo/database/memdb"
	"github.com/dim4egster/qmallgo/ids"
	"github.com/dim4egster/qmallgo/utils/crypto"
	"github.com/dim4egster/qmallgo/utils/units"
	"github.com/dim4egster/qmallgo/vms/avm/fxs"
	"github.com/dim4egster/qmallgo/vms/avm/txs"
	"github.com/dim4egster/qmallgo/vms/components/avax"
	"github.com/dim4egster/qmallgo/vms/nftfx"
	"github.com/dim4egster/qmallgo/vms/propertyfx"
	"github.com/dim4egster/qmallgo/vms/secp256k1fx"
)

var (
	networkID uint32 = 10
	chainID          = ids.ID{5, 4, 3, 2, 1}
	assetID          = ids.ID{1, 2, 3}
	keys             = crypto.BuildTestKeys()
)

func TestTxState(t *testing.T) {
	require := require.New(t)

	db := memdb.New()
	parser, err := txs.NewParser([]fxs.Fx{
		&secp256k1fx.Fx{},
		&nftfx.Fx{},
		&propertyfx.Fx{},
	})
	require.NoError(err)

	stateIntf, err := NewTxState(db, parser, prometheus.NewRegistry())
	require.NoError(err)

	s := stateIntf.(*txState)

	_, err = s.GetTx(ids.Empty)
	require.Equal(database.ErrNotFound, err)

	tx := &txs.Tx{
		Unsigned: &txs.BaseTx{
			BaseTx: avax.BaseTx{
				NetworkID:    networkID,
				BlockchainID: chainID,
				Ins: []*avax.TransferableInput{{
					UTXOID: avax.UTXOID{
						TxID:        ids.Empty,
						OutputIndex: 0,
					},
					Asset: avax.Asset{ID: assetID},
					In: &secp256k1fx.TransferInput{
						Amt: 20 * units.KiloAvax,
						Input: secp256k1fx.Input{
							SigIndices: []uint32{
								0,
							},
						},
					},
				}},
			},
		},
	}

	err = tx.SignSECP256K1Fx(parser.Codec(), [][]*crypto.PrivateKeySECP256K1R{{keys[0]}})
	require.NoError(err)

	err = s.PutTx(ids.Empty, tx)
	require.NoError(err)

	loadedTx, err := s.GetTx(ids.Empty)
	require.NoError(err)
	require.Equal(tx.ID(), loadedTx.ID())

	s.txCache.Flush()

	loadedTx, err = s.GetTx(ids.Empty)
	require.NoError(err)
	require.Equal(tx.ID(), loadedTx.ID())

	err = s.DeleteTx(ids.Empty)
	require.NoError(err)

	_, err = s.GetTx(ids.Empty)
	require.Equal(database.ErrNotFound, err)

	s.txCache.Flush()

	_, err = s.GetTx(ids.Empty)
	require.Equal(database.ErrNotFound, err)
}
