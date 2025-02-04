// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package router

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/dim4egster/qmallgo/api/health"
	"github.com/dim4egster/qmallgo/ids"
	"github.com/dim4egster/qmallgo/message"
	"github.com/dim4egster/qmallgo/snow/networking/benchlist"
	"github.com/dim4egster/qmallgo/snow/networking/handler"
	"github.com/dim4egster/qmallgo/snow/networking/timeout"
	"github.com/dim4egster/qmallgo/utils/logging"
)

// Router routes consensus messages to the Handler of the consensus
// engine that the messages are intended for
type Router interface {
	ExternalHandler
	InternalHandler

	Initialize(
		nodeID ids.NodeID,
		log logging.Logger,
		msgCreator message.InternalMsgBuilder,
		timeouts timeout.Manager,
		shutdownTimeout time.Duration,
		criticalChains ids.Set,
		whiteListedSubnets ids.Set,
		onFatal func(exitCode int),
		healthConfig HealthConfig,
		metricsNamespace string,
		metricsRegisterer prometheus.Registerer,
	) error
	Shutdown()
	AddChain(chain handler.Handler)
	health.Checker
}

// InternalHandler deals with messages internal to this node
type InternalHandler interface {
	benchlist.Benchable

	RegisterRequest(
		nodeID ids.NodeID,
		chainID ids.ID,
		requestID uint32,
		op message.Op,
	)
}
