// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package bootstrap

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"go.uber.org/zap"

	"github.com/dim4egster/qmallgo/ids"
	"github.com/dim4egster/qmallgo/snow/choices"
	"github.com/dim4egster/qmallgo/snow/consensus/snowstorm"
	"github.com/dim4egster/qmallgo/snow/engine/avalanche/vertex"
	"github.com/dim4egster/qmallgo/snow/engine/common/queue"
	"github.com/dim4egster/qmallgo/utils/logging"
)

var errMissingTxDependenciesOnAccept = errors.New("attempting to accept a transaction with missing dependencies")

type txParser struct {
	log                     logging.Logger
	numAccepted, numDropped prometheus.Counter
	vm                      vertex.DAGVM
}

func (p *txParser) Parse(txBytes []byte) (queue.Job, error) {
	tx, err := p.vm.ParseTx(txBytes)
	if err != nil {
		return nil, err
	}
	return &txJob{
		log:         p.log,
		numAccepted: p.numAccepted,
		numDropped:  p.numDropped,
		tx:          tx,
	}, nil
}

type txJob struct {
	log                     logging.Logger
	numAccepted, numDropped prometheus.Counter
	tx                      snowstorm.Tx
}

func (t *txJob) ID() ids.ID { return t.tx.ID() }
func (t *txJob) MissingDependencies() (ids.Set, error) {
	missing := ids.Set{}
	deps, err := t.tx.Dependencies()
	if err != nil {
		return missing, err
	}
	for _, dep := range deps {
		if dep.Status() != choices.Accepted {
			missing.Add(dep.ID())
		}
	}
	return missing, nil
}

// Returns true if this tx job has at least 1 missing dependency
func (t *txJob) HasMissingDependencies() (bool, error) {
	deps, err := t.tx.Dependencies()
	if err != nil {
		return false, err
	}
	for _, dep := range deps {
		if dep.Status() != choices.Accepted {
			return true, nil
		}
	}
	return false, nil
}

func (t *txJob) Execute() error {
	hasMissingDeps, err := t.HasMissingDependencies()
	if err != nil {
		return err
	}
	if hasMissingDeps {
		t.numDropped.Inc()
		return errMissingTxDependenciesOnAccept
	}

	status := t.tx.Status()
	switch status {
	case choices.Unknown, choices.Rejected:
		t.numDropped.Inc()
		return fmt.Errorf("attempting to execute transaction with status %s", status)
	case choices.Processing:
		txID := t.tx.ID()
		if err := t.tx.Verify(); err != nil {
			t.log.Error("transaction failed verification during bootstrapping",
				zap.Stringer("txID", txID),
				zap.Error(err),
			)
			return fmt.Errorf("failed to verify transaction in bootstrapping: %w", err)
		}

		t.numAccepted.Inc()
		t.log.Trace("accepting transaction in bootstrapping",
			zap.Stringer("txID", txID),
		)
		if err := t.tx.Accept(); err != nil {
			t.log.Error("transaction failed to accept during bootstrapping",
				zap.Stringer("txID", txID),
				zap.Error(err),
			)
			return fmt.Errorf("failed to accept transaction in bootstrapping: %w", err)
		}
	}
	return nil
}
func (t *txJob) Bytes() []byte { return t.tx.Bytes() }
