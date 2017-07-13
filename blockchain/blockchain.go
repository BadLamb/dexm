package blockchain

import (
	"time"
	"encoding/binary"

)

type Block struct{
	Index int
	Timestamp int
	PreviousBlockHash string
	TransactionList []byte
	ContractList []byte
}

func (b *Block) CalculateHash(){
	// Convert struct to binary in order to hash
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, b)
	if err != nil {
		log.Error(err)
	}
	return string(blake2b.Sum256(buf))
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
		ContractList: []byte(),
	}
	
	return newBC
}

func (bc *BlockChain) GetLatestBlock() *Block {
	return bc.Blocks[-1]
}

func (bc *BlockChain) NewBlock(transactionList, contractList []byte){
	newB Block
	latestBlock := bc.GetLatestBlock()
	newB.Index = latestBlock.Index + 1
	newB.Timestamp = time.Now().Unix()
	newB.PreviousBlockHash = latestBlock.CalculateHash()
	newB.TransactionList = transactionList
	newB.ContractList = contractList
	
	bc.Blocks = append(bc.Blocks, newB)
}
