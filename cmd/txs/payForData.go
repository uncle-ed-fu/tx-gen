package txs

import (
	"fmt"

	"github.com/celestiaorg/celestia-app/testutil/blobfactory"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/pomifer/tx-gen/cmd/accounts"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
)

func GenPayForDataCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "pfd",
		Short: "generate sdk transactions",
		RunE: func(cmd *cobra.Command, args []string) error {
			cctx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			size, err := cmd.Flags().GetInt("size")
			if err != nil {
				return err
			}

			actors, err := cmd.Flags().GetInt("actors")
			if err != nil {
				return err
			}

			txs := blobfactory.RandBlobTxsWithAccounts(cctx.TxConfig.TxEncoder(), cctx.Keyring, cctx.GRPCClient, size, 1, false, cctx.ChainID, accounts.ActorAccountNames[:actors])

			hashes := make([]string, len(txs))

			for i, tx := range txs {
				res, err := cctx.BroadcastTxSync(tx)
				if err != nil {
					fmt.Println("failure to broadcast", err)
				}

				if abci.CodeTypeOK != res.Code {
					fmt.Println("failed with code", res.Code, res.Logs.String(), res.Data)
					continue
				}
				hashes[i] = res.TxHash
				fmt.Println(res.Height, res.TxHash)
			}

			return nil
		},
	}

	command.Flags().Int("size", 1000, "specify the size of the message")
	command.Flags().Int("actors", 20, "specify the size of the message")

	return command
}
