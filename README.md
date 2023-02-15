

# Daemon Server

# 목차
- [Daemon Server](#daemon-server)
	- [동작 방식](#동작-방식)
	- [stack](#stack)
	- [Getting Start](#getting-start)
- [Block의 정보와 Transaction 정보 가져오기](#block-transaction-정보-가져오기)
	- [ethclient 초기화](#ethclient-초기화)
	- [subscribe](#subscribe)
	- [채널을 이용한 수신](#채널을-이용한-수신)
- [ERC20 Transaction 정보 가져오기](#erc20-transaction-정보-가져오기)
	- [ERC20 결과](#erc20-결과)
- [Contract Address Transaction 가져오기](#contract-address-transaction-가져오기)
	- [Contract Address 결과](#contract-address-결과) 	
	

## 동작 방식
<img width="646" alt="image" src="https://user-images.githubusercontent.com/20445415/218470011-794318af-4199-4444-8c84-923df245d05d.png">

**블록을 참조하여 실제로 블록에 기록된 트랜잭션을 확인**해야 합니다. 이러한 작업을 사용자가 직접적으로 제어하지 않고, 백그라운드에서 돌면서 자동으로 해주는 프로그램을 **블록체인 Daemon(데몬)**이라고 합니다.

블록체인 Daemon으로는 전체 모든 트랜잭션과 블록을 수집할 수도(ex. 블록 익스플로러), 특정한 스마트 컨트랙트 주소(서비스에서 사용하는) 또는 사용자의 지갑 주소를 포함하는 트랜잭션과 블록만을 수집할 수도 있습니다.

## Stack
- ganache
- golang
- mongoDB

## Getting Start

- ganache 실행
  - config.toml 파일에 맞게 포트번호 설정
- mongoDB 실행

# Block Transaction 정보 가져오기

## ethclient 초기화

```go
client, err := ethclient.Dial(cf.Network.URL)
	if err != nil {
		log.Fatal(err)
	}
```

## subscribe

```go
headers := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Fatal(err)
	}
```

## 채널을 이용한 수신

```go
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

> <https://github.com/FDongFDong/blockchain_daemon_server/tree/main/block_trasaction>

___

# ERC20 Transaction 정보 가져오기 
- main.go
    
    ```go
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
    ```
    
    - 트랜잭션에 input data가 포함되어 있으면 함수를 호출할 가능성이 있음을 확인하고
    - input data를 파싱한다.
- erc20/erc20.go
    
    ```go
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
    ```
    
- erc20/erc20.go
    
    ```go
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
    ```
    
    - data의 길이가 136개의 글자수를 가지고 있으면 ERC20의 토큰일 수 있다.
    - 8자리가 “a9059cbb”로 시작하면 ERC20이다.
        - 32~72자리에 있는 데이터는 `To Address` 이다.
            - 해당 주소에서 `transfer(address,uint256)` 를 입력하면 `“a9059cbb”` 가 출력되는 것을 확인할 수 있다.
            
            [Keccak-256](https://emn178.github.io/online-tools/keccak_256.html)
            
        - 72~136 자리는 토큰의`Amount`
            - 앞에 0을 모두 제거한다.
    - value는 계산해서 처리해줘야한다.
- utils/utils.go
    
    ```go
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
    ```
## ERC20 결과

실제 위믹스 Testnet에서의 사용 예시

<img width="1367" alt="image" src="https://user-images.githubusercontent.com/20445415/218989157-6011fd2c-c70d-4b88-a2e6-29e773317a6e.png">

# Contract Address Transaction 가져오기

- main.go
    
    ```go
    if contractAddress, err := utils.GetContractAddress(client, tx.Hash()); err != nil {
    	log.Fatal(err)
    } else {
    	fmt.Println("GetContractAddress : ", contractAddress)
    }
    ```
    
    - GetContractAddress
        - 트랜잭션 해시값을 받는다.
- utils/utils.go
    
    ```go
    func GetContractAddress(client *ethclient.Client, txid common.Hash) (string, error) {
    	receipt_tx, err := client.TransactionReceipt(context.Background(), txid)
    	if err != nil {
    		return "", err
    	}
    	
    	return receipt_tx.ContractAddress.Hex(), nil
    }
    ```
    
    - Receipt에 있는 Contract Address를 가져온다.
    

## Contract Address 결과
- 배포 진행
  
  ![image](https://user-images.githubusercontent.com/20445415/218993799-b7fdf847-2509-428b-a7b7-3468381d2b67.png)
- 결과 
  
  ![image](https://user-images.githubusercontent.com/20445415/218993901-628e0032-1563-4ca4-9960-9bebdd7ed77f.png)
 
> https://github.com/FDongFDong/blockchain_daemon_server/tree/main/erc20_server
