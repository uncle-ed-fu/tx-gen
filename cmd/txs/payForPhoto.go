package txs

import (
	"fmt"

	"github.com/celestiaorg/celestia-app/app"
	"github.com/celestiaorg/celestia-app/x/payment/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	coretypes "github.com/tendermint/tendermint/types"
)

func PayForPhotoCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "photo",
		Short: "generate sdk transactions",
		RunE: func(cmd *cobra.Command, args []string) error {
			cctx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// path, err := cmd.Flags().GetString("path")
			// if err != nil {
			// 	return err
			// }

			// if path == "" {
			// 	return errors.New("empty path, pls use path flag")
			// }

			// file, err := os.Open(path)
			// if err != nil {
			// 	return err
			// }
			// defer file.Close()

			// img, _, err := image.Decode(file)
			// if err != nil {
			// 	return err
			// }

			// buf := bytes.NewBuffer([]byte{})

			// err = jpeg.Encode(buf, img, nil)
			// if err != nil {
			// 	return err
			// }

			// imageBytes := buf.Bytes()
			// copyImageBytes := make([]byte, len(imageBytes))
			// copy(copyImageBytes, imageBytes)

			txs, err := specificSignedWirePayForDataTxs(cctx, cctx.TxConfig, cctx.Keyring, tmrand.Bytes(10000))
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

	command.Flags().String("path", "", "specify the path to the file")

	return command
}

func specificSignedWirePayForDataTxs(clientCtx client.Context, txConfig client.TxConfig, kr keyring.Keyring, b []byte) ([]coretypes.Tx, error) {
	const maxMsgSize = 500000
	chunks := chunkSlice(b, maxMsgSize)
	txs := make([]coretypes.Tx, len(chunks))
	// controlAcc, err := clientCtx.Keyring.Key("actor-97")
	// if err != nil {
	// 	return nil, err
	// }

	// controlAddr, err := controlAcc.GetAddress()
	// if err != nil {
	// 	return nil, err
	// }

	signer := types.NewKeyringSigner(clientCtx.Keyring, "actor-95", clientCtx.ChainID)

	err := signer.UpdateAccountFromClient(clientCtx)
	if err != nil {
		return nil, err
	}

	// _, sequence, err := clientCtx.AccountRetriever.GetAccountNumberSequence(clientCtx, controlAddr)
	// if err != nil {
	// 	return nil, err
	// }

	for i, chunk := range chunks {

		// signer.SetSequence(sequence)

		coin := sdk.Coin{
			Denom:  app.BondDenom,
			Amount: sdk.NewInt(1),
		}

		opts := []types.TxBuilderOption{
			types.SetFeeAmount(sdk.NewCoins(coin)),
			types.SetGasLimit(1000000000),
		}

		fmt.Println("chunk len", len(chunk))

		// create a msg
		msg, err := types.NewWirePayForData(
			[]byte{42, 42, 42, 42, 42, 42, 42, 42},
			chunk,
			types.AllSquareSizes(len(chunk))...,
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
		// sequence++
	}

	return txs, nil
}

func chunkSlice(msg []byte, chunkSize int) [][]byte {
	var chunks [][]byte
	for i := 0; i < len(msg); i += chunkSize {
		end := i + chunkSize

		if end > len(msg) {
			end = len(msg)
		}

		chunks = append(chunks, msg[i:end])
	}

	return chunks
}
