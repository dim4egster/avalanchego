// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package primary

import (
	"context"

	"github.com/dim4egster/qmallgo/ids"
	"github.com/dim4egster/qmallgo/utils/constants"
	"github.com/dim4egster/qmallgo/vms/avm"
	"github.com/dim4egster/qmallgo/vms/platformvm"
	"github.com/dim4egster/qmallgo/vms/platformvm/txs"
	"github.com/dim4egster/qmallgo/vms/secp256k1fx"
	"github.com/dim4egster/qmallgo/wallet/chain/p"
	"github.com/dim4egster/qmallgo/wallet/chain/x"
	"github.com/dim4egster/qmallgo/wallet/subnet/primary/common"
)

var _ Wallet = &wallet{}

// Wallet provides chain wallets for the primary network.
type Wallet interface {
	P() p.Wallet
	X() x.Wallet
}

type wallet struct {
	p p.Wallet
	x x.Wallet
}

func (w *wallet) P() p.Wallet { return w.p }
func (w *wallet) X() x.Wallet { return w.x }

// NewWalletFromURI returns a wallet that supports issuing transactions to the
// chains living in the primary network to a provided [uri].
//
// On creation, the wallet attaches to the provided [uri] and fetches all UTXOs
// that reference any of the keys contained in [kc]. If the UTXOs are modified
// through an external issuance process, such as another instance of the wallet,
// the UTXOs may become out of sync.
//
// The wallet manages all UTXOs locally, and performs all tx signing locally.
func NewWalletFromURI(ctx context.Context, uri string, kc *secp256k1fx.Keychain) (Wallet, error) {
	pCTX, xCTX, utxos, err := FetchState(ctx, uri, kc.Addrs)
	if err != nil {
		return nil, err
	}
	return NewWalletWithState(uri, pCTX, xCTX, utxos, kc), nil
}

// Creates a wallet with pre-loaded/cached P-chain transactions.
func NewWalletWithTxs(ctx context.Context, uri string, kc *secp256k1fx.Keychain, preloadTXs ...ids.ID) (Wallet, error) {
	pCTX, xCTX, utxos, err := FetchState(ctx, uri, kc.Addrs)
	if err != nil {
		return nil, err
	}
	pTXs := make(map[ids.ID]*txs.Tx)
	pClient := platformvm.NewClient(uri)
	for _, id := range preloadTXs {
		txBytes, err := pClient.GetTx(ctx, id)
		if err != nil {
			return nil, err
		}
		tx, err := txs.Parse(txs.Codec, txBytes)
		if err != nil {
			return nil, err
		}
		pTXs[id] = tx
	}
	return NewWalletWithTxsAndState(uri, pCTX, xCTX, utxos, kc, pTXs), nil
}

// Creates a wallet with pre-loaded/cached P-chain transactions and state.
func NewWalletWithTxsAndState(
	uri string,
	pCTX p.Context,
	xCTX x.Context,
	utxos UTXOs,
	kc *secp256k1fx.Keychain,
	pTXs map[ids.ID]*txs.Tx,
) Wallet {
	pUTXOs := NewChainUTXOs(constants.PlatformChainID, utxos)
	pBackend := p.NewBackend(pCTX, pUTXOs, pTXs)
	pBuilder := p.NewBuilder(kc.Addrs, pBackend)
	pSigner := p.NewSigner(kc, pBackend)
	pClient := platformvm.NewClient(uri)

	xChainID := xCTX.BlockchainID()
	xUTXOs := NewChainUTXOs(xChainID, utxos)
	xBackend := x.NewBackend(xCTX, xChainID, xUTXOs)
	xBuilder := x.NewBuilder(kc.Addrs, xBackend)
	xSigner := x.NewSigner(kc, xBackend)
	xClient := avm.NewClient(uri, "X")

	return NewWallet(
		p.NewWallet(pBuilder, pSigner, pClient, pBackend),
		x.NewWallet(xBuilder, xSigner, xClient, xBackend),
	)
}

// Creates a wallet with pre-fetched state.
func NewWalletWithState(
	uri string,
	pCTX p.Context,
	xCTX x.Context,
	utxos UTXOs,
	kc *secp256k1fx.Keychain,
) Wallet {
	pTXs := make(map[ids.ID]*txs.Tx)
	return NewWalletWithTxsAndState(uri, pCTX, xCTX, utxos, kc, pTXs)
}

// Creates a Wallet with the given set of options
func NewWalletWithOptions(w Wallet, options ...common.Option) Wallet {
	return NewWallet(
		p.NewWalletWithOptions(w.P(), options...),
		x.NewWalletWithOptions(w.X(), options...),
	)
}

// Creates a new default wallet
func NewWallet(p p.Wallet, x x.Wallet) Wallet {
	return &wallet{
		p: p,
		x: x,
	}
}
