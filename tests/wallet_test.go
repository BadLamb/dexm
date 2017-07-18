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
    diff := protocol.FindDiff("../v1", "../v2")
    err := diff.Apply("../v1", "../v3")
    if err != nil{
        t.Error(err)
    }
}