// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package sender

import (
	"github.com/dim4egster/qmallgo/ids"
	"github.com/dim4egster/qmallgo/message"
)

// ExternalSender sends consensus messages to other validators
// Right now this is implemented in the networking package
type ExternalSender interface {
	// Send a message to a specific set of nodes
	Send(
		msg message.OutboundMessage,
		nodeIDs ids.NodeIDSet,
		subnetID ids.ID,
		validatorOnly bool,
	) ids.NodeIDSet

	// Send a message to a random group of nodes in a subnet.
	// Nodes are sampled based on their validator status.
	Gossip(
		msg message.OutboundMessage,
		subnetID ids.ID,
		validatorOnly bool,
		numValidatorsToSend int,
		numNonValidatorsToSend int,
		numPeersToSend int,
	) ids.NodeIDSet
}
