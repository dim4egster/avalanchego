// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package nftfx

import (
	"github.com/dim4egster/avalanchego/ids"
	"github.com/dim4egster/avalanchego/snow"
	"github.com/dim4egster/avalanchego/vms"
)

var (
	_ vms.Factory = &Factory{}

	// ID that this Fx uses when labeled
	ID = ids.ID{'n', 'f', 't', 'f', 'x'}
)

type Factory struct{}

func (f *Factory) New(*snow.Context) (interface{}, error) { return &Fx{}, nil }
