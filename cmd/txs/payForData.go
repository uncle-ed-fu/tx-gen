package txs

import (
	"bytes"
	"fmt"

	"github.com/celestiaorg/celestia-app/app"
	"github.com/celestiaorg/celestia-app/x/payment/types"
	"github.com/celestiaorg/nmt/namespace"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pomifer/tx-gen/cmd/accounts"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	"github.com/tendermint/tendermint/pkg/consts"
	coretypes "github.com/tendermint/tendermint/types"
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

			msgSize, err := cmd.Flags().GetInt("size")
			if err != nil {
				return err
			}

			actors, err := cmd.Flags().GetInt("actors")
			if err != nil {
				return err
			}

			txs, err := generateSignedWirePayForDataTxs(cctx, cctx.TxConfig, cctx.Keyring, msgSize, accounts.ActorAccountNames[:actors]...)
			if err != nil {
				return err
			}

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

func generateSignedWirePayForDataTxs(clientCtx client.Context, txConfig client.TxConfig, kr keyring.Keyring, msgSize int, accounts ...string) ([]coretypes.Tx, error) {
	txs := make([]coretypes.Tx, len(accounts))
	for i, account := range accounts {
		signer := types.NewKeyringSigner(kr, account, clientCtx.ChainID)

		err := signer.UpdateAccountFromClient(clientCtx)
		if err != nil {
			return nil, err
		}

		coin := sdk.Coin{
			Denom:  app.BondDenom,
			Amount: sdk.NewInt(1),
		}

		opts := []types.TxBuilderOption{
			types.SetFeeAmount(sdk.NewCoins(coin)),
			types.SetGasLimit(1000000000),
		}

		thisMessageSize := msgSize
		if msgSize < 1 {
			for {
				thisMessageSize = tmrand.NewRand().Intn(100000)
				if thisMessageSize != 0 {
					break
				}
			}
		}

		// create a msg
		msg, err := types.NewWirePayForData(
			randomValidNamespace(),
			tmrand.Bytes(thisMessageSize),
			types.AllSquareSizes(thisMessageSize)...,
		)
		if err != nil {
			return nil, err
		}

		err = msg.SignShareCommitments(signer, opts...)
		if err != nil {
			return nil, err
		}

		builder := signer.NewTxBuilder(opts...)

		tx, err := signer.BuildSignedTx(builder, msg)
		if err != nil {
			return nil, err
		}

		rawTx, err := txConfig.TxEncoder()(tx)
		if err != nil {
			return nil, err
		}

		txs[i] = coretypes.Tx(rawTx)
	}

	return txs, nil
}

func randomValidNamespace() namespace.ID {
	for {
		s := tmrand.Bytes(8)
		if bytes.Compare(s, consts.MaxReservedNamespace) > 0 {
			return s
		}
	}
}
