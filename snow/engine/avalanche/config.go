// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package avalanche

import (
	"github.com/dim4egster/qmallgo/snow"
	"github.com/dim4egster/qmallgo/snow/consensus/avalanche"
	"github.com/dim4egster/qmallgo/snow/engine/avalanche/vertex"
	"github.com/dim4egster/qmallgo/snow/engine/common"
	"github.com/dim4egster/qmallgo/snow/validators"
)

// Config wraps all the parameters needed for an avalanche engine
type Config struct {
	Ctx *snow.ConsensusContext
	common.AllGetsServer
	VM         vertex.DAGVM
	Manager    vertex.Manager
	Sender     common.Sender
	Validators validators.Set

	Params    avalanche.Parameters
	Consensus avalanche.Consensus
}
