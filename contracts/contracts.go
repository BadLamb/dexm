package contracts

import (
	"crypto/ecdsa"
	"crypto/x509"
	"github.com/badlamb/dexm/wallet"
	"log"
	"math/big"

	"gopkg.in/mgo.v2/bson"
)

const (
	SMART_CONTRACT    = 1
	FUNCTION_CONTRACT = 2
	CDN_CONTRACT      = 3
	CDN_BUNDLE        = 4
)

type Contract struct {
	PubKey     []byte    `bson:"p"`
	Type       uint8     `bson:"t"`
	Definition []byte    `bson:"d"`
	SenderSig  [2][]byte `bson:"rs"`
}

type CDNContract struct {
	Hashes        map[string][32]byte `bson:"h"`
	MaxCacheNodes uint16              `bson:"n"`
}

func VerifyContract(c *Contract) (bool, error) {
	r, s := c.SenderSig[0], c.SenderSig[1]
	c.SenderSig = [2][]byte{}

	rb := new(big.Int)
	rb.SetBytes(r)

	sb := new(big.Int)
	sb.SetBytes(s)

	genericPubKey, err := x509.ParsePKIXPublicKey(c.PubKey)
	if err != nil {
		return false, err
	}

	senderPub := genericPubKey.(*ecdsa.PublicKey)

	marshaled, _ := bson.Marshal(c)
	return ecdsa.Verify(senderPub, marshaled, rb, sb), nil
}

func (c *Contract) AppendKeyAndSign(w *wallet.Wallet) {
	x509Encoded, err := x509.MarshalPKIXPublicKey(&w.PrivKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	c.PubKey = x509Encoded

	bsond, _ := bson.Marshal(c)

	// Sign the contract
	r, s := w.Sign(bsond)

	sig := [2][]byte{}
	sig[0] = r.Bytes()
	sig[1] = s.Bytes()

	c.SenderSig = sig
}