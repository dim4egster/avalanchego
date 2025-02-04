// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txs

import (
	"errors"
	"math"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/stretchr/testify/require"

	"github.com/dim4egster/qmallgo/ids"
	"github.com/dim4egster/qmallgo/snow"
	"github.com/dim4egster/qmallgo/utils/constants"
	"github.com/dim4egster/qmallgo/vms/components/avax"
	"github.com/dim4egster/qmallgo/vms/platformvm/fx"
	"github.com/dim4egster/qmallgo/vms/platformvm/validator"
	"github.com/dim4egster/qmallgo/vms/secp256k1fx"
)

func TestAddPermissionlessDelegatorTxSyntacticVerify(t *testing.T) {
	type test struct {
		name   string
		txFunc func(*gomock.Controller) *AddPermissionlessDelegatorTx
		err    error
	}

	var (
		networkID = uint32(1337)
		chainID   = ids.GenerateTestID()
	)

	ctx := &snow.Context{
		ChainID:   chainID,
		NetworkID: networkID,
	}

	// A BaseTx that already passed syntactic verification.
	verifiedBaseTx := BaseTx{
		SyntacticallyVerified: true,
	}

	// A BaseTx that passes syntactic verification.
	validBaseTx := BaseTx{
		BaseTx: avax.BaseTx{
			NetworkID:    networkID,
			BlockchainID: chainID,
		},
	}

	// A BaseTx that fails syntactic verification.
	invalidBaseTx := BaseTx{}

	errCustom := errors.New("custom error")

	tests := []test{
		{
			name: "nil tx",
			txFunc: func(*gomock.Controller) *AddPermissionlessDelegatorTx {
				return nil
			},
			err: ErrNilTx,
		},
		{
			name: "already verified",
			txFunc: func(*gomock.Controller) *AddPermissionlessDelegatorTx {
				return &AddPermissionlessDelegatorTx{
					BaseTx: verifiedBaseTx,
				}
			},
			err: nil,
		},
		{
			name: "no provided stake",
			txFunc: func(*gomock.Controller) *AddPermissionlessDelegatorTx {
				return &AddPermissionlessDelegatorTx{
					BaseTx:    validBaseTx,
					StakeOuts: nil,
				}
			},
			err: errNoStake,
		},
		{
			name: "invalid rewards owner",
			txFunc: func(ctrl *gomock.Controller) *AddPermissionlessDelegatorTx {
				rewardsOwner := fx.NewMockOwner(ctrl)
				rewardsOwner.EXPECT().Verify().Return(errCustom)
				return &AddPermissionlessDelegatorTx{
					BaseTx: validBaseTx,
					Validator: validator.Validator{
						Wght: 1,
					},
					Subnet: ids.GenerateTestID(),
					StakeOuts: []*avax.TransferableOutput{
						{
							Asset: avax.Asset{
								ID: ids.GenerateTestID(),
							},
							Out: &secp256k1fx.TransferOutput{
								Amt: 1,
							},
						},
					},
					DelegationRewardsOwner: rewardsOwner,
				}
			},
			err: errCustom,
		},
		{
			name: "invalid stake output",
			txFunc: func(ctrl *gomock.Controller) *AddPermissionlessDelegatorTx {
				rewardsOwner := fx.NewMockOwner(ctrl)
				rewardsOwner.EXPECT().Verify().Return(nil).AnyTimes()

				stakeOut := avax.NewMockTransferableOut(ctrl)
				stakeOut.EXPECT().Verify().Return(errCustom)
				return &AddPermissionlessDelegatorTx{
					BaseTx: validBaseTx,
					Validator: validator.Validator{
						Wght: 1,
					},
					Subnet: ids.GenerateTestID(),
					StakeOuts: []*avax.TransferableOutput{
						{
							Asset: avax.Asset{
								ID: ids.GenerateTestID(),
							},
							Out: stakeOut,
						},
					},
					DelegationRewardsOwner: rewardsOwner,
				}
			},
			err: errCustom,
		},
		{
			name: "multiple staked assets",
			txFunc: func(ctrl *gomock.Controller) *AddPermissionlessDelegatorTx {
				rewardsOwner := fx.NewMockOwner(ctrl)
				rewardsOwner.EXPECT().Verify().Return(nil).AnyTimes()
				return &AddPermissionlessDelegatorTx{
					BaseTx: validBaseTx,
					Validator: validator.Validator{
						Wght: 1,
					},
					Subnet: ids.GenerateTestID(),
					StakeOuts: []*avax.TransferableOutput{
						{
							Asset: avax.Asset{
								ID: ids.GenerateTestID(),
							},
							Out: &secp256k1fx.TransferOutput{
								Amt: 1,
							},
						},
						{
							Asset: avax.Asset{
								ID: ids.GenerateTestID(),
							},
							Out: &secp256k1fx.TransferOutput{
								Amt: 1,
							},
						},
					},
					DelegationRewardsOwner: rewardsOwner,
				}
			},
			err: errMultipleStakedAssets,
		},
		{
			name: "stake not sorted",
			txFunc: func(ctrl *gomock.Controller) *AddPermissionlessDelegatorTx {
				rewardsOwner := fx.NewMockOwner(ctrl)
				rewardsOwner.EXPECT().Verify().Return(nil).AnyTimes()
				assetID := ids.GenerateTestID()
				return &AddPermissionlessDelegatorTx{
					BaseTx: validBaseTx,
					Validator: validator.Validator{
						Wght: 1,
					},
					Subnet: ids.GenerateTestID(),
					StakeOuts: []*avax.TransferableOutput{
						{
							Asset: avax.Asset{
								ID: assetID,
							},
							Out: &secp256k1fx.TransferOutput{
								Amt: 2,
							},
						},
						{
							Asset: avax.Asset{
								ID: assetID,
							},
							Out: &secp256k1fx.TransferOutput{
								Amt: 1,
							},
						},
					},
					DelegationRewardsOwner: rewardsOwner,
				}
			},
			err: errOutputsNotSorted,
		},
		{
			name: "weight mismatch",
			txFunc: func(ctrl *gomock.Controller) *AddPermissionlessDelegatorTx {
				rewardsOwner := fx.NewMockOwner(ctrl)
				rewardsOwner.EXPECT().Verify().Return(nil).AnyTimes()
				assetID := ids.GenerateTestID()
				return &AddPermissionlessDelegatorTx{
					BaseTx: validBaseTx,
					Validator: validator.Validator{
						Wght: 1,
					},
					Subnet: ids.GenerateTestID(),
					StakeOuts: []*avax.TransferableOutput{
						{
							Asset: avax.Asset{
								ID: assetID,
							},
							Out: &secp256k1fx.TransferOutput{
								Amt: 1,
							},
						},
						{
							Asset: avax.Asset{
								ID: assetID,
							},
							Out: &secp256k1fx.TransferOutput{
								Amt: 1,
							},
						},
					},
					DelegationRewardsOwner: rewardsOwner,
				}
			},
			err: errDelegatorWeightMismatch,
		},
		{
			name: "valid subnet validator",
			txFunc: func(ctrl *gomock.Controller) *AddPermissionlessDelegatorTx {
				rewardsOwner := fx.NewMockOwner(ctrl)
				rewardsOwner.EXPECT().Verify().Return(nil).AnyTimes()
				assetID := ids.GenerateTestID()
				return &AddPermissionlessDelegatorTx{
					BaseTx: validBaseTx,
					Validator: validator.Validator{
						Wght: 2,
					},
					Subnet: ids.GenerateTestID(),
					StakeOuts: []*avax.TransferableOutput{
						{
							Asset: avax.Asset{
								ID: assetID,
							},
							Out: &secp256k1fx.TransferOutput{
								Amt: 1,
							},
						},
						{
							Asset: avax.Asset{
								ID: assetID,
							},
							Out: &secp256k1fx.TransferOutput{
								Amt: 1,
							},
						},
					},
					DelegationRewardsOwner: rewardsOwner,
				}
			},
			err: nil,
		},
		{
			name: "valid primary network validator",
			txFunc: func(ctrl *gomock.Controller) *AddPermissionlessDelegatorTx {
				rewardsOwner := fx.NewMockOwner(ctrl)
				rewardsOwner.EXPECT().Verify().Return(nil).AnyTimes()
				assetID := ids.GenerateTestID()
				return &AddPermissionlessDelegatorTx{
					BaseTx: validBaseTx,
					Validator: validator.Validator{
						Wght: 2,
					},
					Subnet: constants.PrimaryNetworkID,
					StakeOuts: []*avax.TransferableOutput{
						{
							Asset: avax.Asset{
								ID: assetID,
							},
							Out: &secp256k1fx.TransferOutput{
								Amt: 1,
							},
						},
						{
							Asset: avax.Asset{
								ID: assetID,
							},
							Out: &secp256k1fx.TransferOutput{
								Amt: 1,
							},
						},
					},
					DelegationRewardsOwner: rewardsOwner,
				}
			},
			err: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tx := tt.txFunc(ctrl)
			err := tx.SyntacticVerify(ctx)
			require.ErrorIs(err, tt.err)
		})
	}

	t.Run("invalid BaseTx", func(t *testing.T) {
		require := require.New(t)
		tx := &AddPermissionlessDelegatorTx{
			BaseTx: invalidBaseTx,
			Validator: validator.Validator{
				NodeID: ids.GenerateTestNodeID(),
			},
			StakeOuts: []*avax.TransferableOutput{
				{
					Asset: avax.Asset{
						ID: ids.GenerateTestID(),
					},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
		}
		err := tx.SyntacticVerify(ctx)
		require.Error(err)
	})

	t.Run("stake overflow", func(t *testing.T) {
		require := require.New(t)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		rewardsOwner := fx.NewMockOwner(ctrl)
		rewardsOwner.EXPECT().Verify().Return(nil).AnyTimes()
		assetID := ids.GenerateTestID()
		tx := &AddPermissionlessDelegatorTx{
			BaseTx: validBaseTx,
			Validator: validator.Validator{
				NodeID: ids.GenerateTestNodeID(),
				Wght:   1,
			},
			Subnet: ids.GenerateTestID(),
			StakeOuts: []*avax.TransferableOutput{
				{
					Asset: avax.Asset{
						ID: assetID,
					},
					Out: &secp256k1fx.TransferOutput{
						Amt: math.MaxUint64,
					},
				},
				{
					Asset: avax.Asset{
						ID: assetID,
					},
					Out: &secp256k1fx.TransferOutput{
						Amt: 2,
					},
				},
			},
			DelegationRewardsOwner: rewardsOwner,
		}
		err := tx.SyntacticVerify(ctx)
		require.Error(err)
	})
}

func TestAddPermissionlessDelegatorTxNotValidatorTx(t *testing.T) {
	txIntf := any((*AddPermissionlessDelegatorTx)(nil))
	_, ok := txIntf.(ValidatorTx)
	require.False(t, ok)
}
