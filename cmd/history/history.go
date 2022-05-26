package history

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
)

const (
	StartFlag = "start"
	EndFlag   = "end"
	SaveFlag  = "save"
)

func HistoryCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "history",
		Short: "query historical info for some blocks",
	}

	command.PersistentFlags().Int64(StartFlag, 1, "first block to query")
	command.PersistentFlags().Int64(EndFlag, -1, "last block to query")
	command.PersistentFlags().String(SaveFlag, "", "path to save json file to")

	command.AddCommand(BlockTimeCmd())

	return command
}

func readStartEndFlags(cmd *cobra.Command, cctx client.Context) (start, end int64, err error) {
	start, err = cmd.Flags().GetInt64(StartFlag)
	if err != nil {
		return 0, 0, err
	}
	end, err = cmd.Flags().GetInt64(EndFlag)
	if err != nil {
		return 0, 0, err
	}

	if end < 0 {
		res, err := cctx.Client.Header(cmd.Context(), nil)
		if err != nil {
			return 0, 0, err
		}
		end = res.Header.Height
	}
	return
}
