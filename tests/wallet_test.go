package tests

import (
    "github.com/badlamb/DeXm/wallet"
    "os"
    "testing"
)

func TestWallet(t *testing.T) {
    first := wallet.GenerateWallet()
    second := wallet.GenerateWallet()
    if first.GetWallet() == second.GetWallet(){
        t.Error("Wallets don't differ.")
    }

    os.Remove("wallet.pem")
    first.SaveKey("wallet.pem")
    imp := wallet.ImportWallet("wallet.pem")
    os.Remove("wallet.pem")

    if first.GetWallet() != imp.GetWallet(){
        t.Error("Wallet differs across imports")
    }

    imp.Sign([]byte("hh"))
}