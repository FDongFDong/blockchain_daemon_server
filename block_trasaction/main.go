package main

import (
	conf "blockchain_daemon_server/config"
	"blockchain_daemon_server/model"
	"context"
	"fmt"
	"log"
	"math/big"

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

	// subscribe
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

			block, err := client.BlockByHash(context.Background(), header.Hash())
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(block.Hash().Hex())
			fmt.Println(block.Number().Uint64())
			fmt.Println(block.Time())
			fmt.Println(block.Nonce())
			fmt.Println(len(block.Transactions()))

			// TODO: 블록 구조체 생성
			b := model.Block{
				BlockHash:    block.Hash(),
				BlockNumber:  block.Number(),
				GasLimit:     block.GasLimit(),
				GasUsed:      block.GasUsed(),
				Time:         block.Time(),
				Nonce:        block.Nonce(),
				Transactions: make([]model.Transaction, 0),
			}

			// TODO: 트랜잭션 추출
			txs := block.Transactions()
			if len(txs) > 0 {
				for _, tx := range txs {
					t := model.Transaction{}
					t.TxHash = tx.Hash()
					msg, err := tx.AsMessage(types.LatestSignerForChainID(tx.ChainId()), big.NewInt(1))
					if err != nil {
						log.Fatal(err)
					}
					t.From = msg.From()
					t.To = tx.To()
					t.Nonce = tx.Nonce()
					t.GasPrice = tx.GasPrice()
					t.GasLimit = tx.Gas()
					t.Amount = tx.Value()
					t.BlockHash = block.Hash()
					t.BlockNumber = block.Number().Uint64()
					if tx.To() != nil {
						t.To = tx.To()
					}
					b.Transactions = append(b.Transactions, t)
				}
			}

			// TODO: 트랜잭션 구조체 생성

			// DB 저장
			err = md.SaveBlock(&b)
			if err != nil {
				log.Fatal(err)
			}

		}
	}
}
