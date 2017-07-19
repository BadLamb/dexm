package blockchain

import(
    "encoding/json"
	"crypto/x509"
    "crypto/ecdsa"

    "github.com/badlamb/dexm/wallet"
)

// TODO Verify balance and nonce

func VerifyTransaction(encoded []byte) (bool, error) {
    var transaction wallet.Transaction
    err := json.Unmarshal(encoded, &transaction)
    if err != nil{
        return false, err
    }

    r, s := transaction.SenderSig[0], transaction.SenderSig[1]
    transaction.SenderSig = nil

    genericPubKey, err := x509.ParsePKIXPublicKey(transaction.Sender)
    if err != nil{
        return false, err
    }

    senderPub := genericPubKey.(*ecdsa.PublicKey)

    marshaled, _ := json.Marshal(transaction)
    return ecdsa.Verify(senderPub, marshaled, r, s), nil
}