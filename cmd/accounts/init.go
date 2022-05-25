package accounts

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/spf13/cobra"
)

func InitCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "init",
		Short: "generate sdk transactions",
		RunE: func(cmd *cobra.Command, args []string) error {
			cctx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			kb := cctx.Keyring

			for _, acc := range ActorAccountNames {
				rec, _, err := kb.NewMnemonic(acc, keyring.English, "", "", hd.Secp256k1)
				if err != nil {
					return err
				}

				addr, err := rec.GetAddress()
				if err != nil {
					return err
				}

				fmt.Println("generated new account for", addr.String())
			}

			fmt.Println("-----------------------------------------------")

			rec, _, err := kb.NewMnemonic(ContollerAccountName, keyring.English, "", "", hd.Secp256k1)
			if err != nil {
				return err
			}

			addr, err := rec.GetAddress()
			if err != nil {
				return err
			}

			fmt.Println("**controller account**", addr.String())

			return nil
		},
	}

	command.Flags().Int64("amount", 10000, "amount of utia to sent")

	return command
}
