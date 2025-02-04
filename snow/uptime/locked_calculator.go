// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package uptime

import (
	"errors"
	"sync"
	"time"

	"github.com/dim4egster/qmallgo/ids"
	"github.com/dim4egster/qmallgo/utils"
)

var (
	errNotReady = errors.New("should not be called")

	_ LockedCalculator = &lockedCalculator{}
)

type LockedCalculator interface {
	Calculator

	SetCalculator(isBootstrapped *utils.AtomicBool, lock sync.Locker, newC Calculator)
}

type lockedCalculator struct {
	lock           sync.RWMutex
	isBootstrapped *utils.AtomicBool
	calculatorLock sync.Locker
	c              Calculator
}

func NewLockedCalculator() LockedCalculator {
	return &lockedCalculator{}
}

func (c *lockedCalculator) CalculateUptime(nodeID ids.NodeID) (time.Duration, time.Time, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if c.isBootstrapped == nil || !c.isBootstrapped.GetValue() {
		return 0, time.Time{}, errNotReady
	}

	c.calculatorLock.Lock()
	defer c.calculatorLock.Unlock()

	return c.c.CalculateUptime(nodeID)
}

func (c *lockedCalculator) CalculateUptimePercent(nodeID ids.NodeID) (float64, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if c.isBootstrapped == nil || !c.isBootstrapped.GetValue() {
		return 0, errNotReady
	}

	c.calculatorLock.Lock()
	defer c.calculatorLock.Unlock()

	return c.c.CalculateUptimePercent(nodeID)
}

func (c *lockedCalculator) CalculateUptimePercentFrom(nodeID ids.NodeID, startTime time.Time) (float64, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if c.isBootstrapped == nil || !c.isBootstrapped.GetValue() {
		return 0, errNotReady
	}

	c.calculatorLock.Lock()
	defer c.calculatorLock.Unlock()

	return c.c.CalculateUptimePercentFrom(nodeID, startTime)
}

func (c *lockedCalculator) SetCalculator(isBootstrapped *utils.AtomicBool, lock sync.Locker, newC Calculator) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.isBootstrapped = isBootstrapped
	c.calculatorLock = lock
	c.c = newC
}
