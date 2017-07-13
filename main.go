package main

import(
    "os"
    "bufio"
    "strings"

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

                wal.SaveKey(strings.TrimSpace(text))

                return nil
            },
        },

        {
            Name: "startnode",
            Usage: "Starts a full node and tries to sync the blockchain",
            Aliases: []string{"sn"},
            Action: func (c *cli.Context) error{
                protocol.StartSyncServer()
                return nil
            },
        },
    }

    app.Run(os.Args)
}