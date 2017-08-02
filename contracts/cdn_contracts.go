package contracts

import (
	"crypto/x509"
	"io/ioutil"

	"github.com/badlamb/dexm/wallet"
	"github.com/minio/blake2b-simd"
	"gopkg.in/kothar/brotli-go.v0/enc"
	"gopkg.in/mgo.v2/bson"
)

func CreateCDNContract(filenames []string, maxCacheNodes uint16, w *wallet.Wallet) (Contract, error) {
	hashes := make(map[string][32]byte)

	// Make hashes of all files
	for i := 0; i < int(maxCacheNodes); i++ {
		currFile, err := ioutil.ReadFile(filenames[i])
		if err != nil {
			return Contract{}, err
		}

		hash := blake2b.Sum256(currFile)
		hashes[filenames[i]] = hash
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

type CDNFileBundle struct {
	File     []byte    `bson:"fb"`
	Filename string    `bson:"fn"`
	Sig      [2][]byte `bson:"s"`
}

func (c Contract) SelectCDNNodes(w *wallet.Wallet) error {
	var body CDNContract
	err := bson.Unmarshal(c.Definition, body)
	if err != nil {
		return err
	}

	// TODO Randomly select nodes based on proof of burn.

	// Compression settings, this way bundles are smaller
	params := enc.NewBrotliParams()
	params.SetQuality(4)

	for k, _ := range body.Hashes {
		// Compress the file
		rawFile, err := ioutil.ReadFile(k)
		if err != nil {
			return err
		}

		compressedFile := make([]byte, len(rawFile))

		enc.CompressBuffer(params, rawFile, compressedFile)

		thisBundle := CDNFileBundle{
			Filename: k,
			File:     compressedFile,
		}

		byt, err := bson.Marshal(thisBundle)
		if err != nil {
			return err
		}

		// Sign the bundles
		r, s := w.Sign(byt)

		sig := [2][]byte{}
		sig[0] = r.Bytes()
		sig[1] = s.Bytes()

		thisBundle.Sig = sig
	}

	return nil
}
