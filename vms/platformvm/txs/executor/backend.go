// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package executor

import (
	"github.com/dim4egster/qmallgo/snow"
	"github.com/dim4egster/qmallgo/snow/uptime"
	"github.com/dim4egster/qmallgo/utils"
	"github.com/dim4egster/qmallgo/utils/timer/mockable"
	"github.com/dim4egster/qmallgo/vms/platformvm/config"
	"github.com/dim4egster/qmallgo/vms/platformvm/fx"
	"github.com/dim4egster/qmallgo/vms/platformvm/reward"
	"github.com/dim4egster/qmallgo/vms/platformvm/utxo"
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
