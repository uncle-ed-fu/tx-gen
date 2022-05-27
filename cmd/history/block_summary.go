package history

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/spf13/cobra"
)

func BlockSummaryCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "summary",
		Short: "query historical info for some blocks",
		RunE: func(cmd *cobra.Command, args []string) error {
			cctx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			chainID, err := cmd.Flags().GetString(flags.FlagChainID)
			if err != nil {
				return err
			}

			start, end, err := readStartEndFlags(cmd, cctx)
			if err != nil {
				return err
			}

			valMap, err := queryValidators(cmd.Context(), cctx)
			if err != nil {
				return err
			}

			summaries := []BlockSummary{}
			for i := start; i < end; i++ {
				res, err := cctx.Client.Block(cmd.Context(), &i)
				if err != nil {
					return err
				}

				valAddr, err := sdk.ConsAddressFromHex(res.Block.ProposerAddress.String())
				if err != nil {
					return err
				}

				valName, has := valMap[valAddr.String()]
				if !has {
					valName = valAddr.String()
				}

				signers, missedSigners := parseSigners(valMap, res.Block)

				bs := BlockSummary{
					Height:        res.Block.Height,
					Proposer:      valName,
					Signers:       signers,
					MissedSigners: missedSigners,
					Time:          res.Block.Time,
				}

				summaries = append(summaries, bs)

				fmt.Printf("%d - %d - %s\n", res.Block.Height, res.Block.Time.UnixMilli(), valName)
			}

			overallSummary, summaries := SummarizeBlocks(summaries...)
			fmt.Printf(
				"avg %.2f median %.2f stdDev %.2f shortest %.2f longest %.2f\n",
				overallSummary.AverageBlockTime,
				overallSummary.MedianBlockTime,
				overallSummary.StandardDeviationBlockTime,
				overallSummary.ShortestBlockTime,
				overallSummary.LongestBlockTime)

			err = saveJson(fmt.Sprintf("blocks-%d-%d-%s.json", start, end, chainID), summaries)
			if err != nil {
				return err
			}
			return saveJson(fmt.Sprintf("total-summary-%d-%d-%s.json", start, end, chainID), overallSummary)
		},
	}

	return command
}

// returns a map of all the validators where proposer address:name
func queryValidators(ctx context.Context, cctx client.Context) (map[string]string, error) {
	sqc := stakingtypes.NewQueryClient(cctx.GRPCClient)

	valResp, err := sqc.Validators(ctx, &stakingtypes.QueryValidatorsRequest{
		Status: stakingtypes.BondStatusBonded,
		Pagination: &query.PageRequest{
			Offset:     0,
			Limit:      1000,
			CountTotal: true,
		},
	})
	if err != nil {
		return nil, err
	}

	out := make(map[string]string)

	for _, v := range valResp.Validators {
		var pubKey cryptotypes.PubKey
		err = cctx.InterfaceRegistry.UnpackAny(v.ConsensusPubkey, &pubKey)
		if err != nil {
			return nil, err
		}
		conAddr, err := sdk.ConsAddressFromHex(pubKey.Address().String())
		if err != nil {
			return nil, err
		}
		out[conAddr.String()] = v.Description.Moniker
	}
	return out, nil
}
