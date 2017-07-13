package blockchain

import (
	"bytes"
	"encoding/binary"
	"strconv"
	"time"

	"github.com/minio/blake2b-simd"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
)

type Block struct {
	Index             int
	Timestamp         int64
	Hash              string
	PreviousBlockHash string
	TransactionList   []byte
	ContractList      []byte
}

func (b *Block) CalculateHash() string {
	// Convert struct to binary in order to hash
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, b)
	if err != nil {
		log.Error(err)
	}

	hash := blake2b.Sum256(buf.Bytes())
	return string(hash[:])
}

type BlockChain struct {
	DB *leveldb.DB
}

func NewBlockChain() *BlockChain {
	bc := OpenBlockchain()
	// generate Genesis Block
	genesis := Block{
		Timestamp:         time.Now().Unix(),
		PreviousBlockHash: "",
		TransactionList:   []byte("Donald Trump Jr was wrong to meet Russian, says FBI chief Christopher Wray"),
	}

	hash := genesis.CalculateHash()
	genesis.Hash = hash
	bc.DB.Put([]byte(string("len")), []byte(string(1)), nil)
	bc.DB.Put([]byte(string(0)), genesis.GetBytes(), nil)

	return bc
}

func OpenBlockchain() *BlockChain {
	db, err := leveldb.OpenFile("blockchain.db", nil)
	if err != nil {
		log.Fatal(err)
	}

	return &BlockChain{db}
}

func (bc *BlockChain) GetLen() int {
	index, err := bc.DB.Get([]byte(string("len")), nil)
	if err != nil {
		log.Fatal("Invalid blockchain.db")
	}
	num, _ := strconv.Atoi(string(index))
	return num
}

func (bc *BlockChain) GetBlock(index int) (*Block, error) {
	data, err := bc.DB.Get([]byte(string(index)), nil)
	if err != nil {
		return nil, err
	}

	var newBlock Block
	binary.Read(bytes.NewReader(data), binary.LittleEndian, &newBlock)

	return &newBlock, nil
}

func (bc *BlockChain) NewBlock(transactionList, contractList []byte) {
	var newB Block
	lastIndex := bc.GetLen() - 1
	latestBlock, err := bc.GetBlock(lastIndex)
	if err != nil {
		log.Error(err)
		return
	}

	newB.Index = latestBlock.Index + 1
	newB.Timestamp = time.Now().Unix()
	newB.PreviousBlockHash = latestBlock.CalculateHash()
	newB.TransactionList = transactionList
	newB.ContractList = contractList

	bc.DB.Put([]byte(string(lastIndex+1)), newB.GetBytes(), nil)
	bc.DB.Put([]byte(string("len")), []byte(string(lastIndex+2)), nil)
}

func (b *Block) GetBytes() []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, b)
	if err != nil {
		log.Error(err)
	}

	return buf.Bytes()
}
