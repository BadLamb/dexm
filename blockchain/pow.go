package blockchain

import(
    "math/big"

    "github.com/bitgoin/lyra2rev2"
)

// First version of the PoW hash. Uses lyra2rev2
func SumDexmHashVOne(nonce, block []byte) (*big.Int, error){
    toHash := append(nonce, block...)

    result := new(big.Int)

    hash, err := lyra2rev2.Sum(toHash)
    if err != nil{
        return result, err
    }

    result.SetBytes(hash)

    return result, nil
}
