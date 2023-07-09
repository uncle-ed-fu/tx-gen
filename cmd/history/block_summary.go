package history

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	cosmosquery "github.com/cosmos/cosmos-sdk/types/query"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/pomifer/tx-gen/query"
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

			start, end, err := readStartEndFlags(cmd, cctx)
			if err != nil {
				return err
			}

			valMap, err := queryValidators(cmd.Context(), cctx)
			if err != nil {
				return err
			}

			summaries := []query.BlockSummary{}
			var lastTime time.Time
			for i := start; i < end; i++ {
				res, err := cctx.Client.Block(cmd.Context(), &i)
				if err != nil {
					return err
				}

				valAddr := res.Block.ProposerAddress.String()

				valName, has := valMap[valAddr]
				if !has {
					valName = valAddr
				}

				signers, missedSigners := query.ParseSigners(valMap, res.Block)

				bs := query.BlockSummary{
					Height:        res.Block.Height,
					Proposer:      valName,
					Signers:       signers,
					MissedSigners: missedSigners,
					Time:          res.Block.Time,
					Round:         int(res.Block.LastCommit.Round),
				}

				summaries = append(summaries, bs)

				proposer, has := valMap[res.Block.ProposerAddress.String()]
				if !has {
					proposer = res.Block.ProposerAddress.String()
				}
				if proposer == "" {
					proposer = "nil prop" + res.Block.ProposerAddress.String()
				}

				fmt.Printf(
					"%d - rounds %d diff seconds %.2f ### %v\n",
					res.Block.Height,
					res.Block.LastCommit.Round,
					res.Block.Time.Sub(lastTime).Seconds(),
					proposer,
				)
				lastTime = res.Block.Time
			}
			fmt.Println("-----------------------------------------   summary")
			overallSummary, _ := query.SummarizeBlocks(summaries...)
			fmt.Printf(
				"Block times (seconds): avg %.2f stdDev %.2f shortest %.2f longest %.2f \n",
				overallSummary.AverageBlockTime,
				overallSummary.StandardDeviationBlockTime,
				overallSummary.ShortestBlockTime,
				overallSummary.LongestBlockTime,
			)
			fmt.Printf(
				"multi-round heights %d proposer repeats %d \n",
				overallSummary.MultiRoundHeights,
				overallSummary.ProposerRepeats,
			)
			fmt.Printf(
				"percent multi-round heights %.2f\n",
				float64(overallSummary.MultiRoundHeights)/float64(len(summaries)),
			)
			fmt.Println("----------------------------------------- proposer tally")
			for proposer, tally := range overallSummary.ProposerTally {
				fmt.Println(proposer, tally)
			}
			fmt.Println("-----------------------------------------   signer tally", "total", len(summaries))
			for signer, tally := range overallSummary.SignerTally {
				fmt.Println(signer, tally, "missed", len(summaries)-int(tally))
			}
			// err = saveJson(fmt.Sprintf("blocks-%d-%d-%s.json", start, end, chainID), summaries)
			// if err != nil {
			// 	return err
			// }

			// saveJson(fmt.Sprintf("total-summary-%d-%d-%s.json", start, end, chainID), overallSummary)
			return nil
		},
	}

	return command
}

// returns a map of all the validators where proposer address:name
func queryValidators(ctx context.Context, cctx client.Context) (map[string]string, error) {
	sqc := stakingtypes.NewQueryClient(cctx.GRPCClient)

	valResp, err := sqc.Validators(ctx, &stakingtypes.QueryValidatorsRequest{
		Status: stakingtypes.BondStatusBonded,
		Pagination: &cosmosquery.PageRequest{
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
		conAddr := pubKey.Address().String() // sdk.ConsAddressFromHex(pubKey.Address().String())
		out[conAddr] = v.Description.Moniker
	}
	return out, nil
}
