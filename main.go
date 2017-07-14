package main

import(
    "os"
    "bufio"
    "strings"
    "strconv"
    "encoding/json"

    "github.com/urfave/cli"
    log "github.com/sirupsen/logrus"
    "github.com/badlamb/dexm/wallet"
    "github.com/badlamb/dexm/sync"
)

func main(){
    app := cli.NewApp()
    app.Version = "1.0.0 pre-alpha"
    app.Commands = []cli.Command{
        {
            Name: "makewallet",
            Usage: "Generates a new wallet",
            Aliases: []string{"genwallet", "mw", "gw"},
            Action: func (c *cli.Context) error{
                wal := wallet.GenerateWallet()
                log.Info("Generated wallet ", wal.GetWallet())

                log.Info("Input the save path")
                reader := bufio.NewReader(os.Stdin)
                text, _ := reader.ReadString('\n')

                wal.ExportWallet(strings.TrimSpace(text))

                return nil
            },
        },

        {
            Name: "startnode",
            Usage: "Starts a full node and tries to sync the blockchain",
            Aliases: []string{"sn", "rn"},
            Action: func (c *cli.Context) error{
                protocol.StartSyncServer()
                return nil
            },
        },
        
        {
            Name: "maketransaction",
            Usage: "mkt [walletPath] [recipient] [amount]",
            Aliases: []string{"mkt", "gt"},
            Action: func (c *cli.Context) error{
				walletPath := c.Args().Get(0)
                recipient := c.Args().Get(1)
                amount, err := strconv.Atoi(c.Args().Get(2))
                if err != nil {
					log.Error(err)
				}
                senderWallet := wallet.ImportWallet(walletPath)
                transaction, err := senderWallet.NewTransaction(recipient, amount)
                if err != nil{
					log.Error(err)
					return nil
				}
                //the nonce and amount have changed, let's save them
                senderWallet.ExportWallet(walletPath)
                log.Info("Generated Transaction")
                r, _:= json.Marshal(transaction)
                log.Info(string(r))
                return nil
            },
        },
    }

    app.Run(os.Args)
}
