package model

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Model struct {
	client   *mongo.Client
	colBlock *mongo.Collection
}

type Block struct {
	BlockHash    common.Hash   `json:"blockhash" bson:"blockhash"`
	BlockNumber  *big.Int      `json:"blockNumber" bson:"blockNumber"`
	GasLimit     uint64        `json:"gaslimit" bson:"gaslimit"`
	GasUsed      uint64        `json:"gasused" bson:"gasused"`
	Time         uint64        `json:"timestamp" bson:"timestamp"`
	Nonce        uint64        `json:"nonce" bson:"nonce"`
	Transactions []Transaction `json:"transaction" bson:"transaction"`
	// TODO: Block에 대한 구조체를 자유롭게 정의하세요

}

type Transaction struct {
	// TODO: Transaction에 대한 구조체를 자유롭게 정의하세요
	TxHash      common.Hash     `json:"txhash" bson:"txhash"`
	From        common.Address  `json:"from" bson:"from"`
	To          *common.Address `json:"to" bson:"to"`
	Nonce       uint64          `json:"nonce" bson:"nonce"`
	GasPrice    *big.Int        `json:"gas" bson:"gas"`
	GasLimit    uint64          `json:"gaslimit" bson:"gaslimit"`
	Amount      *big.Int        `json:"amount" bson:"amount"`
	BlockHash   common.Hash     `json:"blockhash" bson:"blockhash"`
	BlockNumber uint64          `json:"blocknumber" bson:"blocknumber"`
}

func NewModel(mgUrl string) (*Model, error) {
	r := &Model{}

	var err error
	if r.client, err = mongo.Connect(context.Background(), options.Client().ApplyURI(mgUrl)); err != nil {
		return nil, err
	} else if err := r.client.Ping(context.Background(), nil); err != nil {
		return nil, err
	} else {
		db := r.client.Database("daemon")
		r.colBlock = db.Collection("block")
	}

	return r, nil
}

func (p *Model) SaveBlock(block *Block) error {
	// TODO: Block 데이터를 DB에 저장(생성)하는 함수를 만드세요
	if _, err := p.colBlock.InsertOne(context.TODO(), block); err != nil {
		log.Fatal(err)
		return err
	} else {
		fmt.Println("Insert succeed")
		return nil
	}
}
