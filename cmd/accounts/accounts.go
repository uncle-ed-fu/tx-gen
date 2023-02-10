package accounts

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	NumberOfAccounts     = 100
	ContollerAccountName = "control"
)

var (
	ActorAccountNames = make([]string, NumberOfAccounts)
)

func init() {
	for i := 0; i < NumberOfAccounts; i++ {
		ActorAccountNames[i] = fmt.Sprintf("actor-%d", i)
	}
}

func PrintAccountsCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "printAccounts",
		Short: "print accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(ActorAccountNames)
			return nil
		},
	}
	return command
}
