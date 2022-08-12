// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package executor

import (
	"github.com/dim4egster/avalanchego/snow"
	"github.com/dim4egster/avalanchego/snow/uptime"
	"github.com/dim4egster/avalanchego/utils"
	"github.com/dim4egster/avalanchego/utils/timer/mockable"
	"github.com/dim4egster/avalanchego/vms/platformvm/config"
	"github.com/dim4egster/avalanchego/vms/platformvm/fx"
	"github.com/dim4egster/avalanchego/vms/platformvm/reward"
	"github.com/dim4egster/avalanchego/vms/platformvm/utxo"
)

type Backend struct {
	Config       *config.Config
	Ctx          *snow.Context
	Clk          *mockable.Clock
	Fx           fx.Fx
	FlowChecker  utxo.Verifier
	Uptimes      uptime.Manager
	Rewards      reward.Calculator
	Bootstrapped *utils.AtomicBool
}
