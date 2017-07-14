package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"bytes"
	"hash/crc32"
	"io/ioutil"
	"math/big"
	"time"
	"encoding/binary"
	"encoding/json"

	"github.com/minio/blake2b-simd"
	log "github.com/sirupsen/logrus"
)

type Wallet struct {
	PrivKey *ecdsa.PrivateKey
	Nonce int
	Balance int
}

type WalletFile struct {
	// content to be converted in json
	PrivKeyString string
	Nonce int
	Balance int
}

func GenerateWallet() *Wallet {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	return &Wallet{
		PrivKey: priv,
		Nonce: 0,
		Balance: 0,
	}
}

func ImportWallet(filePath string) *Wallet {
	walletfilejson, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}
	var walletfile WalletFile
	json.Unmarshal(walletfilejson, &walletfile)
	pemEncoded := []byte(walletfile.PrivKeyString)
	decoded, _ := pem.Decode(pemEncoded)

	key, err := x509.ParseECPrivateKey(decoded.Bytes)
	if err != nil {
		log.Fatal(err)
	}
	return &Wallet{
		PrivKey: key,
		Nonce: walletfile.Nonce,
		Balance: walletfile.Balance}
}

func (w *Wallet) SaveKey(filePath string) {
	// convert priv key to x509
	x509Encoded, err := x509.MarshalECPrivateKey(w.PrivKey)
	if err != nil {
		log.Fatal(err)
	}
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "WALLET PRIVATE KEY", Bytes: x509Encoded})
	walletfile := WalletFile{
		PrivKeyString : string(pemEncoded),
		Nonce: w.Nonce,
		Balance: w.Balance,
	}
	
	result, err := json.Marshal(walletfile)
	if err != nil{
		log.Error(err)
	}
	
	ioutil.WriteFile(filePath, result, 400)
}

// The old one was leaking private key data TODO fix
func (w *Wallet) GetWallet() string {
	return BytesToAddress([]byte(w.PrivKey.X.String() + w.PrivKey.Y.String()))
}

func BytesToAddress(data []byte) string {
	hash := blake2b.Sum256(data)
	sum := crc32.ChecksumIEEE(hash[:])
	return fmt.Sprintf("Dexm%s%x", Base58Encoding(hash[:]), sum)
}

func (w *Wallet) Sign(data []byte) (r, s *big.Int) {
	r, s, err := ecdsa.Sign(rand.Reader, w.PrivKey, data)
	if err != nil {
		log.Error(err)
	}

	return r, s
}

// Taken from https://github.com/mr-tron/go-base58
const b58digits_ordered string = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

func Base58Encoding(bin []byte) string {
	binsz := len(bin)
	var i, j, high, zcount, carry int

	for zcount < binsz && bin[zcount] == 0 {
		zcount++
	}

	size := (binsz-zcount)*138/100 + 1
	var buf = make([]byte, size)

	high = size - 1
	for i = zcount; i < binsz; i += 1 {
		j = size - 1
		for carry = int(bin[i]); j > high || carry != 0; j -= 1 {
			carry = carry + 256*int(buf[j])
			buf[j] = byte(carry % 58)
			carry /= 58
		}
		high = j
	}

	for j = 0; j < size && buf[j] == 0; j += 1 {
	}

	var b58 = make([]byte, size-j+zcount)

	if zcount != 0 {
		for i = 0; i < zcount; i++ {
			b58[i] = '1'
		}
	}

	for i = zcount; j < size; i += 1 {
		b58[i] = b58digits_ordered[buf[j]]
		j += 1
	}

	return string(b58)
}

type Transaction struct {
	Sender string
	Recipient string
	
	Amount int
	SenderNonce int
	Timestamp int64
	SenderSig []*big.Int
}

func (w *Wallet) NewTransaction (recipient string, amount int) Transaction{
	var newT Transaction
	newT.Sender = w.GetWallet()
	newT.Recipient = recipient
	newT.Amount = amount
	w.Nonce++
	newT.SenderNonce = w.Nonce
	newT.Timestamp = time.Now().Unix()
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, newT)
	if err != nil {
		log.Error(err)
	}
	r, s := w.Sign(buf.Bytes())
	
	sig := make([]*big.Int, 2)
	sig[0] = r
	sig[1] = s
	
	newT.SenderSig = sig
	return newT
}
