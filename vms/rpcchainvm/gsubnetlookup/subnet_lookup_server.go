// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package gsubnetlookup

import (
	"context"

	"github.com/dim4egster/qmallgo/ids"
	"github.com/dim4egster/qmallgo/snow"

	subnetlookuppb "github.com/dim4egster/qmallgo/proto/pb/subnetlookup"
)

var _ subnetlookuppb.SubnetLookupServer = &Server{}

// Server is a subnet lookup that is managed over RPC.
type Server struct {
	subnetlookuppb.UnsafeSubnetLookupServer
	aliaser snow.SubnetLookup
}

// NewServer returns a subnet lookup connected to a remote subnet lookup
func NewServer(aliaser snow.SubnetLookup) *Server {
	return &Server{aliaser: aliaser}
}

func (s *Server) SubnetID(
	_ context.Context,
	req *subnetlookuppb.SubnetIDRequest,
) (*subnetlookuppb.SubnetIDResponse, error) {
	chainID, err := ids.ToID(req.ChainId)
	if err != nil {
		return nil, err
	}
	id, err := s.aliaser.SubnetID(chainID)
	if err != nil {
		return nil, err
	}
	return &subnetlookuppb.SubnetIDResponse{
		Id: id[:],
	}, nil
}
