package blockchain

import (
	"time"
	"bytes"
	"encoding/binary"

	log "github.com/sirupsen/logrus"
	"github.com/minio/blake2b-simd"
)

type Block struct{
	Index int
	Timestamp int64
	PreviousBlockHash string
	TransactionList []byte
	ContractList []byte
}

func (b *Block) CalculateHash() string{
	// Convert struct to binary in order to hash
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, b)
	if err != nil {
		log.Error(err)
	}
	
	hash := blake2b.Sum256(buf.Bytes())
	return string(hash[:])
}

type BlockChain struct{
	Blocks []Block
}

func NewBlockChain() *BlockChain {
	// generate Genesis Block
	genesis := Block{
		Index: 0,
		Timestamp: time.Now().Unix(),
		PreviousBlockHash: "",
		TransactionList: []byte("Donald Trump Jr was wrong to meet Russian, says FBI chief Christopher Wray"),
	}

	bc := BlockChain{
		Blocks: make([]Block, 1),
	}
	bc.Blocks[0] = genesis
	return &bc
}

func (bc *BlockChain) GetLatestBlock() *Block {
	return &bc.Blocks[len(bc.Blocks) - 1]
}

func (bc *BlockChain) NewBlock(transactionList, contractList []byte){
	var newB Block
	latestBlock := bc.GetLatestBlock()
	newB.Index = latestBlock.Index + 1
	newB.Timestamp = time.Now().Unix()
	newB.PreviousBlockHash = latestBlock.CalculateHash()
	newB.TransactionList = transactionList
	newB.ContractList = contractList
	
	bc.Blocks = append(bc.Blocks, newB)
}
