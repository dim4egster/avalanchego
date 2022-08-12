// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package executor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/dim4egster/avalanchego/ids"
	"github.com/dim4egster/avalanchego/utils/units"
	"github.com/dim4egster/avalanchego/vms/components/avax"
	"github.com/dim4egster/avalanchego/vms/platformvm/state"
	"github.com/dim4egster/avalanchego/vms/platformvm/txs"
	"github.com/dim4egster/avalanchego/vms/secp256k1fx"
)

func TestCreateSubnetTxAP3FeeChange(t *testing.T) {
	ap3Time := defaultGenesisTime.Add(time.Hour)
	tests := []struct {
		name         string
		time         time.Time
		fee          uint64
		expectsError bool
	}{
		{
			name:         "pre-fork - correctly priced",
			time:         defaultGenesisTime,
			fee:          0,
			expectsError: false,
		},
		{
			name:         "post-fork - incorrectly priced",
			time:         ap3Time,
			fee:          100*defaultTxFee - 1*units.NanoAvax,
			expectsError: true,
		},
		{
			name:         "post-fork - correctly priced",
			time:         ap3Time,
			fee:          100 * defaultTxFee,
			expectsError: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			env := newEnvironment()
			env.config.ApricotPhase3Time = ap3Time
			env.ctx.Lock.Lock()
			defer func() {
				assert.NoError(shutdownEnvironment(env))
			}()

			ins, outs, _, signers, err := env.utxosHandler.Spend(preFundedKeys, 0, test.fee, ids.ShortEmpty)
			assert.NoError(err)

			// Create the tx
			utx := &txs.CreateSubnetTx{
				BaseTx: txs.BaseTx{BaseTx: avax.BaseTx{
					NetworkID:    env.ctx.NetworkID,
					BlockchainID: env.ctx.ChainID,
					Ins:          ins,
					Outs:         outs,
				}},
				Owner: &secp256k1fx.OutputOwners{},
			}
			tx := &txs.Tx{Unsigned: utx}
			assert.NoError(tx.Sign(txs.Codec, signers))

			stateDiff, err := state.NewDiff(lastAcceptedID, env)
			assert.NoError(err)

			stateDiff.SetTimestamp(test.time)

			executor := StandardTxExecutor{
				Backend: &env.backend,
				State:   stateDiff,
				Tx:      tx,
			}
			err = tx.Unsigned.Visit(&executor)
			assert.Equal(test.expectsError, err != nil)
		})
	}
}
