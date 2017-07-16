package blockchain

import (
	"encoding/json"
	"time"
	"errors"

	"github.com/minio/blake2b-simd"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
)

type Block struct {
	Index             int64
	Timestamp         int64
	Hash              string
	PreviousBlockHash string
	TransactionList   []byte
	ContractList      []byte
}

func (b *Block) CalculateHash() string {
	// Convert struct to binary in order to hash
	buf := b.GetBytes()

	hash := blake2b.Sum256(buf)
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

	//leveldb has no way to determine the length of the database
	//The key "len" will store that value
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

func (bc *BlockChain) GetLen() int64 {
	size, err :=  bc.DB.SizeOf(nil)
	if err != nil{
		log.Error(err)
		return -1
	}

	return size.Sum()
}

func (bc *BlockChain) GetBlock(index int64) (*Block, error) {
	data, err := bc.DB.Get([]byte(string(index)), nil)
	if err != nil {
		return nil, err
	}

	var newBlock Block
	json.Unmarshal(data, &newBlock)

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

func (bc *BlockChain) VerifyNewBlockValidity(newBlock *Block) (bool, error){
	latestIndex := bc.GetLen() - 1
	latestBlock, err := bc.GetBlock(latestIndex)
	if err != nil {
		return false, err
	}
	if newBlock.Index != latestIndex + 1 {
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
	bCopy.Index = b.Index
	bCopy.Timestamp = b.Timestamp
	bCopy.PreviousBlockHash = b.PreviousBlockHash
	bCopy.TransactionList = b.TransactionList
	bCopy.ContractList = b.ContractList
	
	encoded, err := json.Marshal(bCopy)
	if err != nil{
		log.Error(nil)
		return nil
	}
	log.Debug(string(encoded))
	return encoded
}
