package contracts

import (
	"crypto/x509"
    "crypto/ecdsa"
	"io/ioutil"
    "math/big"

	"github.com/badlamb/dexm/wallet"
	"github.com/minio/blake2b-simd"
	"gopkg.in/mgo.v2/bson"
)

const (
	SMART_CONTRACT    = 1
	FUNCTION_CONTRACT = 2
	CDN_CONTRACT      = 3
)

type Contract struct {
	PubKey     []byte    `bson:"p"`
	Type       uint8     `bson:"t"`
	Definition []byte    `bson:"d"`
	SenderSig  [2][]byte `bson:"rs"`
}

type CDNContract struct {
	Hashes        [][32]byte `bson:"h"`
	MaxCacheNodes uint16     `bson:"n"`
}

func CreateCDNContract(filenames []string, maxCacheNodes uint16, w *wallet.Wallet) (Contract, error) {
	hashes := [][32]byte{}

	// Make hashes of all files
	for i := 0; i < int(maxCacheNodes); i++ {
		currFile, err := ioutil.ReadFile(filenames[i])
		if err != nil {
			return Contract{}, err
		}

		hash := blake2b.Sum256(currFile)
		hashes = append(hashes, hash)
	}

	// Make body of contract
	cdn := CDNContract{
		Hashes:        hashes,
		MaxCacheNodes: maxCacheNodes,
	}

	encoded, err := bson.Marshal(cdn)
	if err != nil {
		return Contract{}, err
	}

	x509Encoded, err := x509.MarshalPKIXPublicKey(w.PrivKey.PublicKey)
	if err != nil {
		return Contract{}, err
	}

	// Put it in the envelope
	toSign := Contract{
		PubKey:     x509Encoded,
		Type:       CDN_CONTRACT,
		Definition: encoded,
	}

	bsond, _ := bson.Marshal(toSign)

	// Sign the contract
	r, s := w.Sign(bsond)

	sig := [2][]byte{}
	sig[0] = r.Bytes()
	sig[1] = s.Bytes()

	toSign.SenderSig = sig

	return toSign, nil
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