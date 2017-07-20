package blockchain

import (
	"errors"
	"time"

	"github.com/minio/blake2b-simd"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"gopkg.in/mgo.v2/bson"
)

type Block struct {
	Index             int64  `bson:"i"`
	Timestamp         int64  `bson:"t"`
	Hash              string `bson:"h"`
	PreviousBlockHash string `bson:"p"`
	TransactionList   []byte `bson:"l,omitempty"`
	ContractList      []byte `bson:"c,omitempty"`
	Miner             string `bson:"m"`
}

func (b *Block) CalculateHash() string {
	// Convert struct to binary in order to hash
	buf := b.GetBytes()

	hash := blake2b.Sum256(buf)
	return string(hash[:])
}

type BlockChain struct {
	DB *leveldb.DB
	Balances *leveldb.DB
}

func NewBlockChain() *BlockChain {
	bc := OpenBlockchain()
	// generate Genesis Block
	genesis := Block{
		Index: 0,
		Timestamp:         time.Now().Unix(),
		TransactionList:   []byte("Donald Trump Jr was wrong to meet Russian, says FBI chief Christopher Wray"),
		Miner: "DexmAohWqVstScKHYntofERjPNFqFo7DWdEr7T15pwQmHiG4e30ed4a6",
	}

	hash := genesis.CalculateHash()
	genesis.Hash = hash

	bc.DB.Put([]byte(string(0)), genesis.GetBytes(), nil)
	bc.GenerateBalanceDB()
	return bc
}

func OpenBlockchain() *BlockChain {
	db, err := leveldb.OpenFile("blockchain.db", nil)
	if err != nil {
		log.Fatal(err)
	}

	bal, err := leveldb.OpenFile("balances.db", nil) 
	return &BlockChain{
		DB: db,
		Balances: bal,
	}
}

func (bc *BlockChain) GetLen() int64 {
	size, err := bc.DB.SizeOf(nil)
	if err != nil {
		log.Error(err)
		return -1
	}

	return size.Sum() + 1
}

func (bc *BlockChain) GetBlock(index int64) (*Block, error) {
	data, err := bc.DB.Get([]byte(string(index)), nil)
	if err != nil {
		return nil, err
	}

	var newBlock Block
	bson.Unmarshal(data, &newBlock)

	return &newBlock, nil
}

func (bc *BlockChain) NewBlock(transactionList, contractList []byte) {
	lastIndex := bc.GetLen() - 1
	latestBlock, err := bc.GetBlock(lastIndex)
	if err != nil {
		log.Error(err)
		return
	}

	newB := Block{
		Index:             latestBlock.Index + 1,
		Timestamp:         time.Now().Unix(),
		PreviousBlockHash: latestBlock.CalculateHash(),
		TransactionList:   transactionList,
		ContractList:      contractList,
	}

	bc.DB.Put([]byte(string(lastIndex+1)), newB.GetBytes(), nil)
}

func (bc *BlockChain) VerifyNewBlockValidity(newBlock *Block) (bool, error) {
	latestIndex := bc.GetLen() - 1
	latestBlock, err := bc.GetBlock(latestIndex)
	if err != nil {
		return false, err
	}
	if newBlock.Index != latestIndex+1 {
		err := errors.New("Block index is not correct")
		return false, err
	} else if latestBlock.Hash != newBlock.PreviousBlockHash {
		err := errors.New("Previous block's hash is not correct")
		return false, err
	} else if newBlock.Hash != newBlock.CalculateHash() {
		err := errors.New("Block hash is not correct")
		return false, err
	} // TODO check if has been correctly mined (i.e.: hash with leading zeros, all trans valid...)

	return true, nil
}

func (b Block) GetBytes() []byte {
	// copy the block without the Hash field
	var bCopy Block
	*&bCopy = *&b
	bCopy.Hash = ""

	encoded, err := bson.Marshal(bCopy)
	if err != nil {
		log.Error(nil)
		return nil
	}

	return encoded
}
