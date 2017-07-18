package schelling

/*
Schelling is a system that enables distributed datastreams on 
the blockchain. It was developed by Vitalik Buterin and you
can read about it here: https://blog.ethereum.org/2014/03/28/schellingcoin-a-minimal-trust-universal-data-feed/
*/

import(
    "encoding/json"
    "sort"

    "github.com/badlamb/dexm/sync"
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

func ComputeValidRange(knownSolutions []int) (min, max int){
    sort.Ints(knownSolutions)

    lowerBound := (len(knownSolutions)/100)*25
    higherBound := (len(knownSolutions)/100)*75

    return knownSolutions[lowerBound], knownSolutions[higherBound]
}

func VerifyHash(hash []byte, wallet string, solution int) bool{
    return protocol.TestEq(hash, ComputeShellingHash(wallet, solution))
}