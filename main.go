package main

import (
	"os"
	"strconv"

	"github.com/badlamb/dexm/blockchain"
	"github.com/badlamb/dexm/sync"
	"github.com/badlamb/dexm/wallet"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"gopkg.in/mgo.v2/bson"
)

func main() {
	app := cli.NewApp()
	app.Version = "1.0.0 pre-alpha"
	app.Commands = []cli.Command{
		{
			Name:    "makewallet",
			Usage:   "mw [filename]",
			Aliases: []string{"genwallet", "mw", "gw"},
			Action: func(c *cli.Context) error {
				wal := wallet.GenerateWallet()
				log.Info("Generated wallet ", wal.GetWallet())

				if c.Args().Get(0) == "" {
					log.Fatal("Invalid filename")
				}

				wal.ExportWallet(c.Args().Get(0))

				return nil
			},
		},

		{
			Name:    "startnode",
			Usage:   "sn",
			Aliases: []string{"sn", "rn"},
			Action: func(c *cli.Context) error {
				protocol.StartSyncServer()
				return nil
			},
		},

		{
			Name:    "maketransaction",
			Usage:   "mkt [walletPath] [recipient] [amount]",
			Aliases: []string{"mkt", "gt"},
			Action: func(c *cli.Context) error {
				walletPath := c.Args().Get(0)
				recipient := c.Args().Get(1)
				amount, err := strconv.Atoi(c.Args().Get(2))
				if err != nil {
					log.Error(err)
				}
				senderWallet := wallet.ImportWallet(walletPath)
				transaction, err := senderWallet.NewTransaction(recipient, amount)
				if err != nil {
					log.Error(err)
					return nil
				}
				//the nonce and amount have changed, let's save them
				senderWallet.ExportWallet(walletPath)
				log.Info("Generated Transaction")
				b, _ := bson.Marshal(transaction)
				protocol.BroadcastMessage(1, b)

				return nil
			},
		},
		{
			Name:    "getbalance",
			Usage:   "gb [wallet]",
			Aliases: []string{"gb", "fb"},
			Action: func(c *cli.Context) error {
				bc := blockchain.OpenBlockchain()
				bal, _ := bc.GetBalance(c.Args().Get(0))

				log.Info("Balance for given wallet is ", bal)

				return nil
			},
		},
	}

	app.Run(os.Args)
}
