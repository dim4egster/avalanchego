// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package platformvm

import (
	"errors"
	"fmt"
	"time"

	"github.com/gorilla/rpc/v2"

	"github.com/prometheus/client_golang/prometheus"

	"go.uber.org/zap"

	"github.com/dim4egster/avalanchego/cache"
	"github.com/dim4egster/avalanchego/codec"
	"github.com/dim4egster/avalanchego/codec/linearcodec"
	"github.com/dim4egster/avalanchego/database"
	"github.com/dim4egster/avalanchego/database/manager"
	"github.com/dim4egster/avalanchego/ids"
	"github.com/dim4egster/avalanchego/snow"
	"github.com/dim4egster/avalanchego/snow/consensus/snowman"
	"github.com/dim4egster/avalanchego/snow/engine/common"
	"github.com/dim4egster/avalanchego/snow/engine/snowman/block"
	"github.com/dim4egster/avalanchego/snow/uptime"
	"github.com/dim4egster/avalanchego/snow/validators"
	"github.com/dim4egster/avalanchego/utils"
	"github.com/dim4egster/avalanchego/utils/constants"
	"github.com/dim4egster/avalanchego/utils/json"
	"github.com/dim4egster/avalanchego/utils/logging"
	"github.com/dim4egster/avalanchego/utils/math"
	"github.com/dim4egster/avalanchego/utils/timer/mockable"
	"github.com/dim4egster/avalanchego/utils/window"
	"github.com/dim4egster/avalanchego/utils/wrappers"
	"github.com/dim4egster/avalanchego/version"
	"github.com/dim4egster/avalanchego/vms/components/avax"
	"github.com/dim4egster/avalanchego/vms/platformvm/api"
	"github.com/dim4egster/avalanchego/vms/platformvm/blocks"
	"github.com/dim4egster/avalanchego/vms/platformvm/fx"
	"github.com/dim4egster/avalanchego/vms/platformvm/metrics"
	"github.com/dim4egster/avalanchego/vms/platformvm/reward"
	"github.com/dim4egster/avalanchego/vms/platformvm/state"
	"github.com/dim4egster/avalanchego/vms/platformvm/txs"
	"github.com/dim4egster/avalanchego/vms/platformvm/txs/builder"
	"github.com/dim4egster/avalanchego/vms/platformvm/txs/mempool"
	"github.com/dim4egster/avalanchego/vms/platformvm/utxo"
	"github.com/dim4egster/avalanchego/vms/secp256k1fx"

	blockexecutor "github.com/dim4egster/avalanchego/vms/platformvm/blocks/executor"
	txexecutor "github.com/dim4egster/avalanchego/vms/platformvm/txs/executor"
)

const (
	validatorSetsCacheSize        = 64
	maxRecentlyAcceptedWindowSize = 256
	recentlyAcceptedWindowTTL     = 5 * time.Minute
)

var (
	_ block.ChainVM    = &VM{}
	_ secp256k1fx.VM   = &VM{}
	_ validators.State = &VM{}

	errWrongCacheType      = errors.New("unexpectedly cached type")
	errMissingValidatorSet = errors.New("missing validator set")
)

type VM struct {
	Factory
	blockBuilder

	metrics.Metrics
	avax.AddressManager
	avax.AtomicUTXOManager
	*network

	// Used to get time. Useful for faking time during tests.
	clock mockable.Clock

	uptimeManager uptime.Manager

	rewards reward.Calculator

	// The context of this vm
	ctx       *snow.Context
	dbManager manager.Manager

	state       state.State
	utxoHandler utxo.Handler

	// ID of the preferred block
	preferred ids.ID

	fx            fx.Fx
	codecRegistry codec.Registry

	// Bootstrapped remembers if this chain has finished bootstrapping or not
	bootstrapped utils.AtomicBool

	// Maps caches for each subnet that is currently whitelisted.
	// Key: Subnet ID
	// Value: cache mapping height -> validator set map
	validatorSetCaches map[ids.ID]cache.Cacher

	// sliding window of blocks that were recently accepted
	recentlyAccepted *window.Window

	txBuilder         builder.TxBuilder
	txExecutorBackend *txexecutor.Backend
	manager           blockexecutor.Manager
}

// Initialize this blockchain.
// [vm.ChainManager] and [vm.vdrMgr] must be set before this function is called.
func (vm *VM) Initialize(
	ctx *snow.Context,
	dbManager manager.Manager,
	genesisBytes []byte,
	upgradeBytes []byte,
	configBytes []byte,
	toEngine chan<- common.Message,
	_ []*common.Fx,
	appSender common.AppSender,
) error {
	ctx.Log.Verbo("initializing platform chain")

	registerer := prometheus.NewRegistry()
	if err := ctx.Metrics.Register(registerer); err != nil {
		return err
	}

	// Initialize metrics as soon as possible
	if err := vm.Metrics.Initialize("", registerer, vm.WhitelistedSubnets); err != nil {
		return err
	}

	vm.ctx = ctx
	vm.dbManager = dbManager

	vm.codecRegistry = linearcodec.NewDefault()
	vm.fx = &secp256k1fx.Fx{}
	if err := vm.fx.Initialize(vm); err != nil {
		return err
	}

	vm.validatorSetCaches = make(map[ids.ID]cache.Cacher)
	vm.recentlyAccepted = window.New(
		window.Config{
			Clock:   &vm.clock,
			MaxSize: maxRecentlyAcceptedWindowSize,
			TTL:     recentlyAcceptedWindowTTL,
		},
	)

	vm.rewards = reward.NewCalculator(vm.RewardConfig)

	var err error
	vm.state, err = state.New(
		vm.dbManager.Current().Database,
		genesisBytes,
		registerer,
		&vm.Config,
		vm.ctx,
		vm.Metrics.LocalStake,
		vm.Metrics.TotalStake,
		vm.rewards,
	)
	if err != nil {
		return err
	}

	vm.AddressManager = avax.NewAddressManager(ctx)
	vm.AtomicUTXOManager = avax.NewAtomicUTXOManager(ctx.SharedMemory, txs.Codec)
	vm.utxoHandler = utxo.NewHandler(vm.ctx, &vm.clock, vm.state, vm.fx)
	vm.uptimeManager = uptime.NewManager(vm.state)
	vm.UptimeLockedCalculator.SetCalculator(&vm.bootstrapped, &ctx.Lock, vm.uptimeManager)

	vm.txBuilder = builder.NewTxBuilder(
		vm.ctx,
		vm.Config,
		&vm.clock,
		vm.fx,
		vm.state,
		vm.AtomicUTXOManager,
		vm.utxoHandler,
	)

	vm.txExecutorBackend = &txexecutor.Backend{
		Config:       &vm.Config,
		Ctx:          vm.ctx,
		Clk:          &vm.clock,
		Fx:           vm.fx,
		FlowChecker:  vm.utxoHandler,
		Uptimes:      vm.uptimeManager,
		Rewards:      vm.rewards,
		Bootstrapped: &vm.bootstrapped,
	}

	// Note: There is a circular dependency between the mempool and block
	//       builder which is broken by passing in the vm.
	mempool, err := mempool.NewMempool("mempool", registerer, vm)
	if err != nil {
		return fmt.Errorf("failed to create mempool: %w", err)
	}

	vm.manager = blockexecutor.NewManager(
		mempool,
		vm.Metrics,
		vm.state,
		vm.txExecutorBackend,
		vm.recentlyAccepted,
	)

	vm.blockBuilder.Initialize(mempool, vm, toEngine)

	vm.network = newNetwork(appSender, vm)

	if err := vm.updateValidators(); err != nil {
		return fmt.Errorf(
			"failed to initialize validator sets: %w",
			err,
		)
	}

	// Create all of the chains that the database says exist
	if err := vm.initBlockchains(); err != nil {
		return fmt.Errorf(
			"failed to initialize blockchains: %w",
			err,
		)
	}

	lastAcceptedID := vm.state.GetLastAccepted()
	ctx.Log.Info("initializing last accepted",
		zap.Stringer("blkID", lastAcceptedID),
	)
	return vm.SetPreference(lastAcceptedID)
}

// Create all chains that exist that this node validates.
func (vm *VM) initBlockchains() error {
	if err := vm.createSubnet(constants.PrimaryNetworkID); err != nil {
		return err
	}

	if vm.StakingEnabled {
		for subnetID := range vm.WhitelistedSubnets {
			if err := vm.createSubnet(subnetID); err != nil {
				return err
			}
		}
	} else {
		subnets, err := vm.state.GetSubnets()
		if err != nil {
			return err
		}
		for _, subnet := range subnets {
			if err := vm.createSubnet(subnet.ID()); err != nil {
				return err
			}
		}
	}
	return nil
}

// Create the subnet with ID [subnetID]
func (vm *VM) createSubnet(subnetID ids.ID) error {
	chains, err := vm.state.GetChains(subnetID)
	if err != nil {
		return err
	}
	for _, chain := range chains {
		tx, ok := chain.Unsigned.(*txs.CreateChainTx)
		if !ok {
			return fmt.Errorf("expected tx type *txs.CreateChainTx but got %T", chain.Unsigned)
		}
		vm.Config.CreateChain(chain.ID(), tx)
	}
	return nil
}

// onBootstrapStarted marks this VM as bootstrapping
func (vm *VM) onBootstrapStarted() error {
	vm.bootstrapped.SetValue(false)
	return vm.fx.Bootstrapping()
}

// onNormalOperationsStarted marks this VM as bootstrapped
func (vm *VM) onNormalOperationsStarted() error {
	if vm.bootstrapped.GetValue() {
		return nil
	}
	vm.bootstrapped.SetValue(true)

	if err := vm.fx.Bootstrapped(); err != nil {
		return err
	}

	primaryValidatorSet, exist := vm.Validators.GetValidators(constants.PrimaryNetworkID)
	if !exist {
		return errNoPrimaryValidators
	}
	primaryValidators := primaryValidatorSet.List()

	validatorIDs := make([]ids.NodeID, len(primaryValidators))
	for i, vdr := range primaryValidators {
		validatorIDs[i] = vdr.ID()
	}

	if err := vm.uptimeManager.StartTracking(validatorIDs); err != nil {
		return err
	}
	return vm.state.Commit()
}

func (vm *VM) SetState(state snow.State) error {
	switch state {
	case snow.Bootstrapping:
		return vm.onBootstrapStarted()
	case snow.NormalOp:
		return vm.onNormalOperationsStarted()
	default:
		return snow.ErrUnknownState
	}
}

// Shutdown this blockchain
func (vm *VM) Shutdown() error {
	if vm.dbManager == nil {
		return nil
	}

	vm.blockBuilder.Shutdown()

	if vm.bootstrapped.GetValue() {
		primaryValidatorSet, exist := vm.Validators.GetValidators(constants.PrimaryNetworkID)
		if !exist {
			return errNoPrimaryValidators
		}
		primaryValidators := primaryValidatorSet.List()

		validatorIDs := make([]ids.NodeID, len(primaryValidators))
		for i, vdr := range primaryValidators {
			validatorIDs[i] = vdr.ID()
		}

		if err := vm.uptimeManager.Shutdown(validatorIDs); err != nil {
			return err
		}
		if err := vm.state.Commit(); err != nil {
			return err
		}
	}

	errs := wrappers.Errs{}
	errs.Add(
		vm.state.Close(),
		vm.dbManager.Close(),
	)
	return errs.Err
}

// BuildBlock builds a block to be added to consensus
func (vm *VM) BuildBlock() (snowman.Block, error) { return vm.blockBuilder.BuildBlock() }

func (vm *VM) ParseBlock(b []byte) (snowman.Block, error) {
	// Note: blocks to be parsed are not verified, so we must used blocks.Codec
	// rather than blocks.GenesisCodec
	statelessBlk, err := blocks.Parse(blocks.Codec, b)
	if err != nil {
		return nil, err
	}
	return vm.manager.NewBlock(statelessBlk), nil
}

func (vm *VM) GetBlock(blkID ids.ID) (snowman.Block, error) {
	return vm.manager.GetBlock(blkID)
}

// LastAccepted returns the block most recently accepted
func (vm *VM) LastAccepted() (ids.ID, error) {
	return vm.manager.LastAccepted(), nil
}

// SetPreference sets the preferred block to be the one with ID [blkID]
func (vm *VM) SetPreference(blkID ids.ID) error {
	if blkID == vm.preferred {
		// If the preference didn't change, then this is a noop
		return nil
	}
	vm.preferred = blkID
	vm.blockBuilder.ResetBlockTimer()
	return nil
}

func (vm *VM) Preferred() (snowman.Block, error) {
	return vm.manager.GetBlock(vm.preferred)
}

func (vm *VM) Version() (string, error) {
	return version.Current.String(), nil
}

// CreateHandlers returns a map where:
// * keys are API endpoint extensions
// * values are API handlers
func (vm *VM) CreateHandlers() (map[string]*common.HTTPHandler, error) {
	server := rpc.NewServer()
	server.RegisterCodec(json.NewCodec(), "application/json")
	server.RegisterCodec(json.NewCodec(), "application/json;charset=UTF-8")
	server.RegisterInterceptFunc(vm.Metrics.APIRequestMetrics.InterceptRequest)
	server.RegisterAfterFunc(vm.Metrics.APIRequestMetrics.AfterRequest)
	if err := server.RegisterService(&Service{vm: vm}, "platform"); err != nil {
		return nil, err
	}

	return map[string]*common.HTTPHandler{
		"": {
			Handler: server,
		},
	}, nil
}

// CreateStaticHandlers returns a map where:
// * keys are API endpoint extensions
// * values are API handlers
func (vm *VM) CreateStaticHandlers() (map[string]*common.HTTPHandler, error) {
	server := rpc.NewServer()
	server.RegisterCodec(json.NewCodec(), "application/json")
	server.RegisterCodec(json.NewCodec(), "application/json;charset=UTF-8")
	if err := server.RegisterService(&api.StaticService{}, "platform"); err != nil {
		return nil, err
	}

	return map[string]*common.HTTPHandler{
		"": {
			LockOptions: common.NoLock,
			Handler:     server,
		},
	}, nil
}

func (vm *VM) Connected(vdrID ids.NodeID, _ *version.Application) error {
	return vm.uptimeManager.Connect(vdrID)
}

func (vm *VM) Disconnected(vdrID ids.NodeID) error {
	if err := vm.uptimeManager.Disconnect(vdrID); err != nil {
		return err
	}
	return vm.state.Commit()
}

// GetValidatorSet returns the validator set at the specified height for the
// provided subnetID.
func (vm *VM) GetValidatorSet(height uint64, subnetID ids.ID) (map[ids.NodeID]uint64, error) {
	validatorSetsCache, exists := vm.validatorSetCaches[subnetID]
	if !exists {
		validatorSetsCache = &cache.LRU{Size: validatorSetsCacheSize}
		// Only cache whitelisted subnets
		if vm.WhitelistedSubnets.Contains(subnetID) || subnetID == constants.PrimaryNetworkID {
			vm.validatorSetCaches[subnetID] = validatorSetsCache
		}
	}

	if validatorSetIntf, ok := validatorSetsCache.Get(height); ok {
		validatorSet, ok := validatorSetIntf.(map[ids.NodeID]uint64)
		if !ok {
			return nil, errWrongCacheType
		}
		vm.Metrics.ValidatorSetsCached.Inc()
		return validatorSet, nil
	}

	lastAcceptedHeight, err := vm.GetCurrentHeight()
	if err != nil {
		return nil, err
	}
	if lastAcceptedHeight < height {
		return nil, database.ErrNotFound
	}

	// get the start time to track metrics
	startTime := vm.Clock().Time()

	currentValidators, ok := vm.Validators.GetValidators(subnetID)
	if !ok {
		return nil, errMissingValidatorSet
	}
	currentValidatorList := currentValidators.List()

	vdrSet := make(map[ids.NodeID]uint64, len(currentValidatorList))
	for _, vdr := range currentValidatorList {
		vdrSet[vdr.ID()] = vdr.Weight()
	}

	for i := lastAcceptedHeight; i > height; i-- {
		diffs, err := vm.state.GetValidatorWeightDiffs(i, subnetID)
		if err != nil {
			return nil, err
		}

		for nodeID, diff := range diffs {
			var op func(uint64, uint64) (uint64, error)
			if diff.Decrease {
				// The validator's weight was decreased at this block, so in the
				// prior block it was higher.
				op = math.Add64
			} else {
				// The validator's weight was increased at this block, so in the
				// prior block it was lower.
				op = math.Sub64
			}

			newWeight, err := op(vdrSet[nodeID], diff.Amount)
			if err != nil {
				return nil, err
			}
			if newWeight == 0 {
				delete(vdrSet, nodeID)
			} else {
				vdrSet[nodeID] = newWeight
			}
		}
	}

	// cache the validator set
	validatorSetsCache.Put(height, vdrSet)

	endTime := vm.Clock().Time()
	vm.Metrics.ValidatorSetsCreated.Inc()
	vm.Metrics.ValidatorSetsDuration.Add(float64(endTime.Sub(startTime)))
	vm.Metrics.ValidatorSetsHeightDiff.Add(float64(lastAcceptedHeight - height))
	return vdrSet, nil
}

// GetMinimumHeight returns the height of the most recent block beyond the
// horizon of our recentlyAccepted window.
//
// Because the time between blocks is arbitrary, we're only guaranteed that
// the window's configured TTL amount of time has passed once an element
// expires from the window.
//
// To try to always return a block older than the window's TTL, we return the
// parent of the oldest element in the window (as an expired element is always
// guaranteed to be sufficiently stale). If we haven't expired an element yet
// in the case of a process restart, we default to the lastAccepted block's
// height which is likely (but not guaranteed) to also be older than the
// window's configured TTL.
func (vm *VM) GetMinimumHeight() (uint64, error) {
	oldest, ok := vm.recentlyAccepted.Oldest()
	if !ok {
		return vm.GetCurrentHeight()
	}

	blk, err := vm.GetBlock(oldest.(ids.ID))
	if err != nil {
		return 0, err
	}

	return blk.Height() - 1, nil
}

// GetCurrentHeight returns the height of the last accepted block
func (vm *VM) GetCurrentHeight() (uint64, error) {
	lastAccepted, err := vm.GetBlock(vm.state.GetLastAccepted())
	if err != nil {
		return 0, err
	}
	return lastAccepted.Height(), nil
}

func (vm *VM) updateValidators() error {
	primaryValidators, err := vm.state.ValidatorSet(constants.PrimaryNetworkID)
	if err != nil {
		return err
	}
	if err := vm.Validators.Set(constants.PrimaryNetworkID, primaryValidators); err != nil {
		return err
	}

	weight, _ := primaryValidators.GetWeight(vm.ctx.NodeID)
	vm.LocalStake.Set(float64(weight))
	vm.TotalStake.Set(float64(primaryValidators.Weight()))

	for subnetID := range vm.WhitelistedSubnets {
		subnetValidators, err := vm.state.ValidatorSet(subnetID)
		if err != nil {
			return err
		}
		if err := vm.Validators.Set(subnetID, subnetValidators); err != nil {
			return err
		}
	}
	return nil
}

func (vm *VM) CodecRegistry() codec.Registry { return vm.codecRegistry }

func (vm *VM) Clock() *mockable.Clock { return &vm.clock }

func (vm *VM) Logger() logging.Logger { return vm.ctx.Log }

// Returns the percentage of the total stake of the subnet connected to this
// node.
func (vm *VM) getPercentConnected(subnetID ids.ID) (float64, error) {
	vdrSet, exists := vm.Validators.GetValidators(subnetID)
	if !exists {
		return 0, errNoValidators
	}

	vdrSetWeight := vdrSet.Weight()
	if vdrSetWeight == 0 {
		return 1, nil
	}

	var (
		connectedStake uint64
		err            error
	)
	for _, vdr := range vdrSet.List() {
		if !vm.uptimeManager.IsConnected(vdr.ID()) {
			continue // not connected to us --> don't include
		}
		connectedStake, err = math.Add64(connectedStake, vdr.Weight())
		if err != nil {
			return 0, err
		}
	}
	return float64(connectedStake) / float64(vdrSetWeight), nil
}
