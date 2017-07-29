package tests

import (
    "os"
    "testing"

    "github.com/badlamb/dexm/wallet"
    "github.com/badlamb/dexm/sync"
)

func TestWallet(t *testing.T) {
    first := wallet.GenerateWallet()
    second := wallet.GenerateWallet()
    if first.GetWallet() == second.GetWallet(){
        t.Error("Wallets don't differ.")
    }

    os.Remove("wallet.pem")
    first.ExportWallet("wallet.pem")
    imp := wallet.ImportWallet("wallet.pem")
    os.Remove("wallet.pem")

    if first.GetWallet() != imp.GetWallet(){
        t.Error("Wallet differs across imports")
    }

    imp.Sign([]byte("hh"))
}

func TestHotpatch(t *testing.T) {
    diff := protocol.FindDiff("../.testfiles/v1", "../.testfiles/v2")
    diff.Apply("../.testfiles/v1", "../.testfiles/v3")
}