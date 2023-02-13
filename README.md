

# Daemon Server

## 동작 방식
<img width="646" alt="image" src="https://user-images.githubusercontent.com/20445415/218470011-794318af-4199-4444-8c84-923df245d05d.png">

**블록을 참조하여 실제로 블록에 기록된 트랜잭션을 확인**해야 합니다. 이러한 작업을 사용자가 직접적으로 제어하지 않고, 백그라운드에서 돌면서 자동으로 해주는 프로그램을 **블록체인 Daemon(데몬)**이라고 합니다.

블록체인 Daemon으로는 전체 모든 트랜잭션과 블록을 수집할 수도(ex. 블록 익스플로러), 특정한 스마트 컨트랙트 주소(서비스에서 사용하는) 또는 사용자의 지갑 주소를 포함하는 트랜잭션과 블록만을 수집할 수도 있습니다.

## Blockchain Network에서 생성된 Block의 정보와 Transaction 정보 가져오기

## stack
- ganache
- golang
- mongoDB

## 기능 설명

### ethclient 초기화

```solidity
client, err := ethclient.Dial(cf.Network.URL)
	if err != nil {
		log.Fatal(err)
	}
```

### subscribe

```solidity
headers := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Fatal(err)
	}
```

### 채널을 이용한 수신

```solidity
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
					msg, err := tx.AsMessage(types.LatestSignerForChainID(tx.ChainId()), block.BaseFee())
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
				}
			}

			// DB insert
			err = md.SaveBlock(&b)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
```
## 소스코드
