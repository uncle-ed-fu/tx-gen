package main

import (
	"context"
	"log"
	"os"

	"github.com/celestiaorg/celestia-app/app"
	"github.com/celestiaorg/celestia-app/app/encoding"
	"github.com/pomifer/tx-gen/cmd/accounts"
	"github.com/pomifer/tx-gen/cmd/fund"
	"github.com/pomifer/tx-gen/cmd/txs"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/spf13/cobra"
)

func main() {
	defaultNodeHome := "/home/evan/.tx-gen"

	encodingConfig := encoding.MakeEncodingConfig(app.ModuleEncodingRegisters...)

	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount(app.Bech32PrefixAccAddr, app.Bech32PrefixAccPub)
	cfg.SetBech32PrefixForValidator(app.Bech32PrefixValAddr, app.Bech32PrefixValPub)
	cfg.SetBech32PrefixForConsensusNode(app.Bech32PrefixConsAddr, app.Bech32PrefixConsPub)
	cfg.Seal()

	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Codec).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastBlock).
		WithHomeDir(defaultNodeHome).
		WithViper("TX-GEN")

	rootCmd := &cobra.Command{
		Use:   "tx-gen",
		Short: "generate sdk transactions",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			initClientCtx, err := client.ReadPersistentCommandFlags(initClientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			initClientCtx, err = config.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}

			rpcClient, err := client.NewClientFromNode("tcp://localhost:26657")
			if err != nil {
				log.Fatal(err)
			}

			initClientCtx = initClientCtx.WithClient(rpcClient)
			// WithKeyring(kr)

			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			return nil
		},
	}

	rootCmd.PersistentFlags().String(flags.FlagNode, "tcp://localhost:26657", "")
	rootCmd.PersistentFlags().String(flags.FlagChainID, "mamaki", "")

	rootCmd.PersistentFlags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	rootCmd.PersistentFlags().String(flags.FlagKeyringDir, defaultNodeHome, "The client Keyring directory; if omitted, the default 'home' directory will be used")
	rootCmd.PersistentFlags().String(flags.FlagKeyringBackend, "file", "Select keyring's backend (os|file|test)")
	rootCmd.PersistentFlags().String(cli.OutputFlag, "text", "Output format (text|json)")

	rootCmd.AddCommand(
		accounts.InitCmd(),
		fund.FundCmd(),
		txs.GenPayForDataCmd(),
		keys.Commands("/home/evan/.tx-gen/"),
		txs.PayForPhotoCmd(),
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &client.Context{})

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.Fatal(err)
	}
}
