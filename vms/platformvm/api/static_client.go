// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"context"

	"github.com/dim4egster/qmallgo/utils/rpc"
)

var _ StaticClient = &staticClient{}

// StaticClient for interacting with the platformvm static api
type StaticClient interface {
	BuildGenesis(
		ctx context.Context,
		args *BuildGenesisArgs,
		options ...rpc.Option,
	) (*BuildGenesisReply, error)
}

// staticClient is an implementation of a platformvm client for interacting with
// the platformvm static api
type staticClient struct {
	requester rpc.EndpointRequester
}

// NewClient returns a platformvm client for interacting with the platformvm static api
func NewStaticClient(uri string) StaticClient {
	return &staticClient{requester: rpc.NewEndpointRequester(
		uri+"/ext/vm/platform",
		"platform",
	)}
}

func (c *staticClient) BuildGenesis(
	ctx context.Context,
	args *BuildGenesisArgs,
	options ...rpc.Option,
) (resp *BuildGenesisReply, err error) {
	resp = &BuildGenesisReply{}
	err = c.requester.SendRequest(ctx, "buildGenesis", args, resp, options...)
	return resp, err
}
