package blockchain

import (
	"crypto/ecdsa"
	"crypto/x509"
	"errors"
	"math/big"

	"github.com/badlamb/dexm/wallet"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
)

type WalletInfo struct {
	Balance int
	Nonce   int
}

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

func (bc *BlockChain) GetBalance(wallet string) (int, int) {
	val, err := bc.Balances.Get([]byte(wallet), nil)
	if err != nil {
		log.Error(err)
		return 0, 0
	}

	var curr WalletInfo
	err = bson.Unmarshal(val, &curr)

	if err != nil {
		log.Error(err)
		return 0, 0
	}

	return curr.Balance, curr.Nonce
}

func (bc *BlockChain) SetBalance(wallet string, amount, nonce int) error {
	c := WalletInfo{
		Balance: amount,
		Nonce:   nonce,
	}

	data, err := bson.Marshal(c)
	if err != nil {
		return err
	}

	return bc.Balances.Put([]byte(wallet), data, nil)
}

func VerifyTransactionSignature(transaction wallet.Transaction) (bool, error) {
	r, s := transaction.SenderSig[0], transaction.SenderSig[1]
	transaction.SenderSig = [2][]byte{}

	rb := new(big.Int)
	rb.SetBytes(r)

	sb := new(big.Int)
	sb.SetBytes(s)

	genericPubKey, err := x509.ParsePKIXPublicKey(transaction.Sender)
	if err != nil {
		return false, err
	}

	senderPub := genericPubKey.(*ecdsa.PublicKey)

	marshaled, _ := bson.Marshal(transaction)
	return ecdsa.Verify(senderPub, marshaled, rb, sb), nil
}

func (bc *BlockChain) ProcessBlock(curr *Block) error {
	var totalGas = 0

	// Genesis node isn't a valid transaction
	if curr.Index != 0 {
		var transactions []wallet.Transaction
		err := bson.Unmarshal(curr.TransactionList, &transactions)
		if err != nil {
			return err
		}

		for k, v := range transactions {
			status, err := VerifyTransactionSignature(v)
			if err != nil {
				return err
			}

			if !status {
				return errors.New("Invalid signature")
			}

			sender := wallet.BytesToAddress(v.Sender)
			balance, nonce := bc.GetBalance(sender)

			if v.Amount+v.Gas <= balance {
                bc.SetBalance(sender, balance-(v.Amount+v.Gas), nonce+1)

				// As there was no new transaction on the recivers part the nonce doesn't change
				rbal, rnonce := bc.GetBalance(v.Recipient)
				bc.SetBalance(v.Recipient, rbal+v.Amount, rnonce)
			} else {
				return errors.New("Transaction is invalid " + string(k))
			}
		}
	}

	// Give the reward for having mined the block.
	bal, nonce := bc.GetBalance(curr.Miner)

	bc.SetBalance(curr.Miner, bal+GetReward(5)+totalGas, nonce)

	return nil
}
