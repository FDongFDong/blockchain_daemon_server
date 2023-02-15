package erc20

import (
	// "erc20_server/contracts"

	"blockchain_daemon_server/erc20_server/contracts"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func ERC20Transaction(data string) (string, string) {
	// ERC20 토큰은 136개의 글자수로 이루어져 있다.
	// a9059cbb0000000000000000000000004ebbd4881a45b836bac17ea52f1bcef72b787b0e00000000000000000000000000000000000000000000010f0cf064dd59200000

	if len(data) != 136 {
		return "", "0"
	} else {
		// 앞 8자리는 methodID
		methodID := data[:8]
		// 32~72는 to Address
		to := data[32:72]
		// 72~136은 토큰 양
		value := data[72:136]
		if methodID != "a9059cbb" {
			return "", "0"
		}
		i := new(big.Int)
		// 앞에 0 모두 제거
		valueStr := strings.TrimLeft(value, "0")
		i.SetString(valueStr, 16)
		return to, i.String()
	}
}

func GetContractInfo(client *ethclient.Client, to *common.Address) (string, string, uint8) {
	instance, err := contracts.NewContracts(*to, client)
	if err != nil {
		log.Fatal(err)
	}
	name, err := instance.Name(&bind.CallOpts{})
	if err != nil {
		log.Fatal(err)
	}
	symbol, err := instance.Symbol(&bind.CallOpts{})
	if err != nil {
		log.Fatal(err)
	}
	decimals, err := instance.Decimals(&bind.CallOpts{})
	if err != nil {
		log.Fatal(err)
	}
	return name, symbol, decimals

}
