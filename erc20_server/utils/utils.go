package utils

import (
	"context"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func GetContractAddress(client *ethclient.Client, txid common.Hash) (string, error) {
	receipt_tx, err := client.TransactionReceipt(context.Background(), txid)
	if err != nil {
		return "", err
	}
	return receipt_tx.ContractAddress.Hex(), nil
}

func GetRealGasUsed(client *ethclient.Client, txid common.Hash) (uint64, error) {
	receipt_tx, err := client.TransactionReceipt(context.Background(), txid)
	if err != nil {
		return 0, err
	}
	return receipt_tx.GasUsed, nil
}

func GetRealGasPrice(baseFee uint64, maxFeeCap uint64, maxTipCap uint64) *big.Int {
	if baseFee+maxTipCap > maxFeeCap {
		return new(big.Int).SetUint64(maxFeeCap)
	} else {
		return new(big.Int).SetUint64(baseFee + maxTipCap)
	}
}

func GetRealValue(value string, decimal uint8) string {
	n := new(big.Int)
	pw := int(math.Pow(float64(10), float64(decimal)))
	i := new(big.Int).SetUint64(uint64(pw))
	n, ok := n.SetString(value, 10)
	if !ok {
		return ""
	} else {
		result := big.NewInt(0)
		result.Div(n, i)
		return result.String()
	}
}
