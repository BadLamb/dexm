package contracts

import(
	"gopkg.in/kothar/brotli-go.v0/enc"
)

type SmartContract struct{
    IsWasm bool
    Code []byte
}

func NewSmartContract(code []byte, isWasm bool) (SmartContract, error) {
    params := enc.NewBrotliParams()
	params.SetQuality(4)

    compressedFile, err := enc.CompressBuffer(params, code, make([]byte, 0))
    if err != nil{
        return SmartContract{}, err
    }

    return SmartContract{
        IsWasm: isWasm,
        Code: compressedFile,
    }, nil
}

// TODO Implement the v8 binding and sandbox
// This functions executes a smart contract and return the gas that it needed to run
// func (c SmartContract) Execute() int64{
// }
