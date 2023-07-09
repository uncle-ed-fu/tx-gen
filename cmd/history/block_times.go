package history

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
)

func BlockTimesCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "times",
		Short: "query historical times for some blocks",
		RunE: func(cmd *cobra.Command, args []string) error {
			cctx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			validators, err := queryValidators(cmd.Context(), cctx)
			if err != nil {
				return err
			}

			fmt.Println("validators", validators)

			start, end, err := readStartEndFlags(cmd, cctx)
			if err != nil {
				return err
			}

			var lastTime time.Time
			for i := start; i < end; i++ {
				res, err := cctx.Client.Block(cmd.Context(), &i)
				if err != nil {
					return err
				}

				proposer, has := validators[res.Block.ProposerAddress.String()]
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

			return nil
		},
	}

	return command
}
