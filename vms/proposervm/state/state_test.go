// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package state

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/stretchr/testify/require"

	"github.com/dim4egster/qmallgo/database/memdb"
	"github.com/dim4egster/qmallgo/database/versiondb"
)

func TestState(t *testing.T) {
	a := require.New(t)

	db := memdb.New()
	vdb := versiondb.New(db)
	s := New(vdb)

	testBlockState(a, s)
	testChainState(a, s)
}

func TestMeteredState(t *testing.T) {
	a := require.New(t)

	db := memdb.New()
	vdb := versiondb.New(db)
	s, err := NewMetered(vdb, "", prometheus.NewRegistry())
	a.NoError(err)

	testBlockState(a, s)
	testChainState(a, s)
}
