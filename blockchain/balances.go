package blockchain

import (
	"strconv"

	"github.com/badlamb/dexm/wallet"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
)

func (bc *BlockChain) GenerateBalanceDB() {
	len := bc.GetLen()

	// Iterate all blocks
	for i := int64(0); i < len; i++ {
		curr, err := bc.GetBlock(i)
		if err != nil {
			log.Error(err)
			return
		}

        bc.ProcessBlock(curr)
	}
}

func (bc *BlockChain) GetBalance(wallet string) int {
	val, err := bc.Balances.Get([]byte(wallet), nil)
	if err != nil {
		return 0
	}

	bal, err := strconv.Atoi(string(val))
	if err != nil {
		log.Error(err)
		return 0
	}

	return bal
}

func (bc *BlockChain) SetBalance(wallet string, amount int) error {
	return bc.Balances.Put([]byte(wallet), []byte(string(amount)), nil)
}

func (bc *BlockChain) ProcessBlock(curr *Block) {
	var transactions []wallet.Transaction
	err := bson.Unmarshal(curr.TransactionList, &transactions)
	if err != nil {
		log.Error(err)
		return
	}

	for _, v := range transactions {
		sender := wallet.BytesToAddress(v.Sender)
		balance := bc.GetBalance(sender)

		// TODO Add gas prices
		if v.Amount <= balance {
			bc.SetBalance(sender, balance-v.Amount)
			bc.SetBalance(v.Recipient, bc.GetBalance(v.Recipient)+v.Amount)
		}
	}

	// Give the reward for having mined the block.
	bc.Balances.Put([]byte(curr.Miner), []byte(string(5)), nil)
}
