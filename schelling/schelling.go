package schelling

import(
    "encoding/json"

    "github.com/minio/blake2b-simd"
)

type ShellingSolution struct{
    Value int
    Address string
}

func ComputeShellingHash(wallet string, value int) []byte{
    sol := ShellingSolution{
        Value: value,
        Address: wallet,
    }

    encoded, _ := json.Marshal(sol)
    sum := blake2b.Sum256(encoded)
    
    return sum[:]
}
