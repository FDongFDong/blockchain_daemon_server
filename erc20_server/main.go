package main

import (
	"context"
	"encoding/hex"

	conf "blockchain_daemon_server/erc20_server/config"
	"blockchain_daemon_server/erc20_server/erc20"
	"blockchain_daemon_server/erc20_server/model"
	"blockchain_daemon_server/erc20_server/utils"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	// config 초기화
	cf := conf.GetConfig("./config/config.toml")

	// model 초기화
	md, err := model.NewModel(cf.DB.Host)
	if err != nil {
		log.Fatal(err)
	}

	// ethclint 초기화
	client, err := ethclient.Dial(cf.Network.URL)
	if err != nil {
		log.Fatal(err)
	}

	headers := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case header := <-headers:
			fmt.Println(header.Hash().Hex())

			block, err := client.BlockByNumber(context.Background(), header.Number)
			if err != nil {
				log.Fatal(err)
			}
			baseFee := block.BaseFee()
			// 블록 구조체 생성
			b := model.Block{
				BlockHash:    block.Hash().Hex(),
				BlockNumber:  block.Number().Uint64(),
				GasLimit:     block.GasLimit(),
				GasUsed:      block.GasUsed(),
				Time:         block.Time(),
				Nonce:        block.Nonce(),
				Transactions: make([]model.Transaction, 0),
			}

			// 트랜잭션 추출
			txs := block.Transactions()
			if len(txs) > 0 {
				for _, tx := range txs {
					msg, err := tx.AsMessage(types.LatestSignerForChainID(tx.ChainId()), baseFee)
					if err != nil {
						log.Fatal(err)
					}
					// 트랜잭션 구조체 생성
					t := model.Transaction{
						TxHash:      tx.Hash().Hex(),
						To:          "", // 디폴트 값 처리
						From:        msg.From().Hex(),
						Nonce:       tx.Nonce(),
						GasPrice:    tx.GasPrice().Uint64(),
						GasLimit:    tx.Gas(),
						Amount:      tx.Value().Uint64(),
						BlockHash:   block.Hash().Hex(),
						BlockNumber: block.Number().Uint64(),
					}

					if tx.To() != nil {
						t.To = tx.To().Hex()
					}

					b.Transactions = append(b.Transactions, t)
					if contractAddress, err := utils.GetContractAddress(client, tx.Hash()); err != nil {
						log.Fatal(err)
					} else {
						fmt.Println("GetContractAddress : ", contractAddress)
					}
					if realGasLimit, err := utils.GetRealGasUsed(client, tx.Hash()); err != nil {
						log.Fatal(err)
					} else {
						fmt.Println("realGasLimit : ", realGasLimit)
					}
					realGasPrice := utils.GetRealGasPrice(baseFee.Uint64(), tx.GasFeeCap().Uint64(), tx.GasTipCap().Uint64())
					fmt.Println("realGasPrice :", realGasPrice)
					// input Data에 데이터가 있으면 함수를 호출할 가능성이 있다.
					if len(tx.Data()) != 0 {
						// 실제 ERC20 토큰은 input datad안에 들어있다.
						to, value := erc20.ERC20Transaction(hex.EncodeToString(tx.Data()))
						if to != "" {
							symbol, name, decimal := erc20.GetContractInfo(client, tx.To())
							fmt.Println("ERC20 Contract Address: ", tx.To().Hex())
							fmt.Println("ERC20 Contract to Name: ", name)
							fmt.Println("ERC20 Contract to Symbol: ", symbol)
							fmt.Println("ERC20 Contract to Decimals: ", decimal)
							fmt.Println("ERC20 Transfer to Address: ", to)
							if tokenValue := utils.GetRealValue(value, decimal); tokenValue != "" {
								fmt.Println("ERC20 value :", tokenValue)
							}
						}
					}
				}
			}

			fmt.Println("========================")
			fmt.Println("BlockHash: ", b.BlockHash)
			fmt.Println("BlockNumber: ", b.BlockNumber)
			fmt.Println("GasLimit: ", b.GasLimit)
			fmt.Println("GasUsed: ", b.GasUsed)
			fmt.Println("Time: ", b.Time)
			fmt.Println("Nonce: ", b.Nonce)
			fmt.Println("Transactions: ", b.Transactions)

			fmt.Println("========================")
			// DB insert
			err = md.SaveBlock(&b)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
