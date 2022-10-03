// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package peer

import (
	"time"

	"github.com/dim4egster/qmallgo/ids"
	"github.com/dim4egster/qmallgo/message"
	"github.com/dim4egster/qmallgo/network/throttling"
	"github.com/dim4egster/qmallgo/snow/networking/router"
	"github.com/dim4egster/qmallgo/snow/networking/tracker"
	"github.com/dim4egster/qmallgo/snow/validators"
	"github.com/dim4egster/qmallgo/utils/logging"
	"github.com/dim4egster/qmallgo/utils/timer/mockable"
	"github.com/dim4egster/qmallgo/version"
)

type Config struct {
	// Size, in bytes, of the buffer this peer reads messages into
	ReadBufferSize int
	// Size, in bytes, of the buffer this peer writes messages into
	WriteBufferSize         int
	Clock                   mockable.Clock
	Metrics                 *Metrics
	MessageCreator          message.Creator
	MessageCreatorWithProto message.Creator

	// TODO: remove this once we complete blueberry migration
	BlueberryTime time.Time

	Log                  logging.Logger
	InboundMsgThrottler  throttling.InboundMsgThrottler
	Network              Network
	Router               router.InboundHandler
	VersionCompatibility version.Compatibility
	MySubnets            ids.Set
	Beacons              validators.Set
	NetworkID            uint32
	PingFrequency        time.Duration
	PongTimeout          time.Duration
	MaxClockDifference   time.Duration

	// Unix time of the last message sent and received respectively
	// Must only be accessed atomically
	LastSent, LastReceived int64

	// Tracks CPU/disk usage caused by each peer.
	ResourceTracker tracker.ResourceTracker
}

func (c *Config) GetMessageCreator() message.Creator {
	now := c.Clock.Time()
	if c.IsBlueberryActivated(now) {
		return c.MessageCreatorWithProto
	}
	return c.MessageCreator
}

func (c *Config) IsBlueberryActivated(time time.Time) bool {
	return !time.Before(c.BlueberryTime)
}
