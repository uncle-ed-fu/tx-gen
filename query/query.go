package query

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmosquery "github.com/cosmos/cosmos-sdk/types/query"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// returns a map of all the validators where proposer address:name
func QueryValidators(ctx context.Context, cctx client.Context) (map[string]string, error) {
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
		conAddr, err := sdk.ConsAddressFromHex(pubKey.Address().String())
		if err != nil {
			return nil, err
		}
		out[conAddr.String()] = v.Description.Moniker
	}
	return out, nil
}
