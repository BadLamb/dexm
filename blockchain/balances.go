package blockchain

import (
	"crypto/rand"
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
	Burn int
}

// Generates database of all balances in the blockchain
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

// Given a wallet returns balance and nonce
func (bc *BlockChain) GetBalance(wallet string) (int, int, int) {
	val, err := bc.Balances.Get([]byte(wallet), nil)
	if err != nil {
		log.Error(err)
		return 0, 0, 0
	}

	var curr WalletInfo
	err = bson.Unmarshal(val, &curr)

	if err != nil {
		log.Error(err)
		return 0, 0, 0
	}

	return curr.Balance, curr.Nonce, curr.Nonce
}

// Stores amount, nonce, and burn for a given wallet 
func (bc *BlockChain) SetBalance(wallet string, amount, nonce, burn int) error {
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

// Verify if a transaction has a valid signature
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

// Takes in a block and then updates all balances
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
			balance, nonce, burn := bc.GetBalance(sender)

			// Gas is capped at 1% of the amount. TODO Make gas based on transaction size not
			// amount, this way all transaction have the same importance to the network.
			if v.Amount/100 <= v.Gas{
				return errors.New("Too much gas!")
			}

			// Check if balance is enough to complete the transaction
			if v.Amount+v.Gas <= balance {
				// Check if the transaction is for the Proof of burn addr, if it is then add burn
				if v.Recipient == "DexmProofOfBurn"{
					burn += v.Amount
				}

                bc.SetBalance(sender, balance-(v.Amount+v.Gas), nonce+1, burn)

                totalGas += v.Gas

				// As there was no new transaction on the recivers part the nonce doesn't change
				rbal, rnonce, rburn := bc.GetBalance(v.Recipient)
				bc.SetBalance(v.Recipient, rbal+v.Amount, rnonce, rburn)
			} else {
				return errors.New("Transaction is invalid " + string(k))
			}
		}
	}

	// Give the reward for having mined the block.
	bal, nonce, burn := bc.GetBalance(curr.Miner)

	bc.SetBalance(curr.Miner, bal+GetReward(5)+totalGas, nonce, burn)

	return nil
}

// Returns random wallets based on their burn. 
// TODO Add PoB decay
func (bc *BlockChain) GetPoBWallets(nodes int) []string{
	iter := bc.Balances.NewIterator(nil, nil)
	balances := make(map[int]string)
	
	var totalBal int64

	// Get all the balances in memory
	for iter.Next() {
		var CurrWall WalletInfo
		bson.Unmarshal(iter.Value(), &CurrWall)

		balances[CurrWall.Burn] = string(iter.Key())

		totalBal += int64(CurrWall.Burn)
	}

	b := big.NewInt(totalBal)

	result := []string{}
	for i := 0; i<nodes; i++{
		// Pick a random burn coin in all of the burnt coins.
		res, _ := rand.Int(rand.Reader, b)

		burnIndex := new(big.Int)

		for k, v := range balances{
			// burnIndex <= res
			status := burnIndex.Cmp(res)
			if status == -1 || status == 0{
				result = append(result, v)
				delete(balances, k)
				break
			}else{
				burnIndex.Add(burnIndex, big.NewInt(int64(k)))
			}
		}
	}

	return result
}