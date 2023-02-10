package fund

import (
	"fmt"

	"github.com/celestiaorg/celestia-app/app"
	"github.com/celestiaorg/celestia-app/x/blob/types"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/pomifer/tx-gen/cmd/accounts"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
	coretypes "github.com/tendermint/tendermint/types"
)

func FundCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "fund",
		Short: "generate sdk transactions",
		RunE: func(cmd *cobra.Command, args []string) error {
			cctx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			amount, err := cmd.Flags().GetInt64("amount")
			if err != nil {
				return err
			}

			count, err := cmd.Flags().GetInt("count")
			if err != nil {
				return err
			}

			txs, err := generateSignedSendTxs(cctx, count, amount, accounts.ActorAccountNames...)
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

				fmt.Println(res.TxHash)

				hashes[i] = res.TxHash
			}

			return nil
		},
	}

	command.Flags().Int64("amount", 100000, "amount of utia to send to each recipient")
	command.Flags().Int("count", 1, "number of times to repeat")

	return command
}

func generateSignedSendTxs(clientCtx client.Context, count int, amount int64, rs ...string) ([]coretypes.Tx, error) {
	txs := make([]coretypes.Tx, len(rs)*count)

	controlAcc, err := clientCtx.Keyring.Key(accounts.ContollerAccountName)
	if err != nil {
		return nil, err
	}

	controlAddr, err := controlAcc.GetAddress()
	if err != nil {
		return nil, err
	}

	signer := types.NewKeyringSigner(clientCtx.Keyring, accounts.ContollerAccountName, clientCtx.ChainID)

	err = signer.UpdateAccountFromClient(clientCtx)
	if err != nil {
		return nil, err
	}

	_, sequence, err := clientCtx.AccountRetriever.GetAccountNumberSequence(clientCtx, controlAddr)
	if err != nil {
		return nil, err
	}

	for j := 0; j < count; j++ {
		for i, account := range rs {
			r, err := clientCtx.Keyring.Key(account)
			if err != nil {
				return nil, err
			}

			rAddr, err := r.GetAddress()
			if err != nil {
				return nil, err
			}

			feeCoin := sdk.Coin{
				Denom:  app.BondDenom,
				Amount: sdk.NewInt(1000000),
			}

			opts := []types.TxBuilderOption{
				types.SetFeeAmount(sdk.NewCoins(feeCoin)),
				types.SetGasLimit(1000000000),
			}

			amountCoin := sdk.Coin{
				Denom:  app.BondDenom,
				Amount: sdk.NewInt(amount),
			}

			builder := signer.NewTxBuilder(opts...)

			msg := banktypes.NewMsgSend(controlAddr, rAddr, sdk.NewCoins(amountCoin))

			tx, err := signer.BuildSignedTx(builder, msg)
			if err != nil {
				return nil, err
			}

			rawTx, err := clientCtx.TxConfig.TxEncoder()(tx)
			if err != nil {
				return nil, err
			}

			index := j * len(rs)

			txs[index+i] = coretypes.Tx(rawTx)

			sequence++
			signer.SetSequence(sequence)
		}
	}

	return txs, nil
}
