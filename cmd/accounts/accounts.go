package accounts

import (
	"fmt"
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
