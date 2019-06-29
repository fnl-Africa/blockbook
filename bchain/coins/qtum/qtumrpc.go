package qtum

import (
	"blockbook/bchain"
	"blockbook/bchain/coins/btc"
	"encoding/json"
	"math/big"

	"github.com/golang/glog"
)

// QtumRPC is an interface to JSON-RPC bitcoind service.
type QtumRPC struct {
	*btc.BitcoinRPC
	minFeeRate *big.Int // satoshi per kb
}

// NewQtumRPC returns new QtumRPC instance.
func NewQtumRPC(config json.RawMessage, pushHandler func(bchain.NotificationType)) (bchain.BlockChain, error) {
	b, err := btc.NewBitcoinRPC(config, pushHandler)
	if err != nil {
		return nil, err
	}

	s := &QtumRPC{
		b.(*btc.BitcoinRPC),
		big.NewInt(400000),
	}
	s.RPCMarshaler = btc.JSONMarshalerV1{}
	s.ChainConfig.SupportsEstimateSmartFee = true

	return s, nil
}

// Initialize initializes QtumRPC instance.
func (b *QtumRPC) Initialize() error {
	ci, err := b.GetChainInfo()
	if err != nil {
		return err
	}
	chainName := ci.Chain

	params := GetChainParams(chainName)

	// always create parser
	b.Parser = NewQtumParser(params, b.ChainConfig)

	// parameters for getInfo request
	if params.Net == MainnetMagic {
		b.Testnet = false
		b.Network = "livenet"
	} else {
		b.Testnet = true
		b.Network = "testnet"
	}

	glog.Info("rpc: block chain ", params.Name)

	return nil
}

// GetTransactionForMempool returns a transaction by the transaction ID
// It could be optimized for mempool, i.e. without block time and confirmations
func (b *QtumRPC) GetTransactionForMempool(txid string) (*bchain.Tx, error) {
	return b.GetTransaction(txid)
}

// EstimateSmartFee returns fee estimation
func (b *QtumRPC) EstimateSmartFee(blocks int, conservative bool) (big.Int, error) {
	feeRate, err := b.BitcoinRPC.EstimateSmartFee(blocks, conservative)
	if err == nil {
		// fix for trustwallet
		newFeeRate := *big.NewInt(feeRate.Int64() * 1024 / 1000)
		return newFeeRate, err
	}
	return feeRate, err
}

// EstimateFee returns fee estimation.
func (b *QtumRPC) EstimateFee(blocks int) (big.Int, error) {
	feeRate, err := b.BitcoinRPC.EstimateFee(blocks)
	if err == nil {
		// fix for trustwallet
		newFeeRate := *big.NewInt(feeRate.Int64() * 1024 / 1000)
		return newFeeRate, err
	}
	return feeRate, err
}
