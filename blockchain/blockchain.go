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

const (
	GENESIS_DIFF = 20000000000000
	USD_REWARD   = 250
)

func (b *Block) CalculateHash() string {
	// Convert struct to binary in order to hash
	buf := b.GetBytes()

	hash := blake2b.Sum256(buf)
	return string(hash[:])
}

type BlockChain struct {
	DB       *leveldb.DB
	Balances *leveldb.DB
}

func NewBlockChain() *BlockChain {
	bc := OpenBlockchain()
	// generate Genesis Block
	genesis := Block{
		Index:           0,
		Timestamp:       time.Now().Unix(),
		TransactionList: []byte("Donald Trump Jr was wrong to meet Russian, says FBI chief Christopher Wray"),
		Miner:           "DexmMmBKy5zsLEC3JK82FwGdFK53d7ae974ca8",
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
		DB:       db,
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

func (b *Block) GetBytes() []byte {
	// copy the block without the Hash field
	var bCopy Block
	*&bCopy = *b
	bCopy.Hash = ""

	encoded, err := bson.Marshal(bCopy)
	if err != nil {
		log.Error(nil)
		return nil
	}

	return encoded
}

/*
This function assumes the block is valid.
TODO Implement adjustments based on Shelling results and hashing power
*/
func (b *Block) GetDifficulty(bc *BlockChain) int64 {
	if b.Index == 0 {
		return GENESIS_DIFF
	}

	prevBlock, err := bc.GetBlock(b.Index - 1)
	if err != nil {
		log.Fatal(err)
	}

	return prevBlock.GetDifficulty(bc)
}

/*
Each block has a fixed reward in USD. usdPrice is found by
uding schelling. We do this to keep the price of the coin somewhat stable.
*/
func GetReward(usdPrice int) int {
	return USD_REWARD / usdPrice
}
