package blockchain

import (
	"crypto/ecdsa"
	"crypto/x509"
	"math/big"

	"github.com/badlamb/dexm/wallet"
	"gopkg.in/mgo.v2/bson"
)

// TODO Verify balance and nonce

func VerifyTransaction(encoded []byte) (bool, error) {
	var transaction wallet.Transaction
	err := bson.Unmarshal(encoded, &transaction)
	if err != nil {
		return false, err
	}

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
