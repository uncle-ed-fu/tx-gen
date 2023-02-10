package profile

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"
)

func ProfileCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "profile",
		Short: "query historical info for some blocks",
		RunE: func(cmd *cobra.Command, args []string) error {
			cctx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			file, err := os.OpenFile("peer_data.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
			if err != nil {
				return err
			}
			defer file.Close()

			var lastStatusResp *rpctypes.ResultStatus
			var lastNetInfo *rpctypes.ResultNetInfo
			for {
				ctx, _ := context.WithTimeout(cmd.Context(), time.Second*1)
				netInfo, err := cctx.Client.NetInfo(ctx)
				if err != nil {
					log.Println(err)
				}
				if netInfo == nil {
					netInfo = lastNetInfo
				}

				statusResp, err := cctx.Client.Status(ctx)
				if err != nil {
					log.Println(err)
				}
				if statusResp == nil {
					statusResp = lastStatusResp
				}

				d := PeerData{
					PeerCount: netInfo.NPeers,
					Height:    statusResp.SyncInfo.LatestBlockHeight,
					LocalTime: time.Now(),
				}

				rawData, err := json.Marshal(d)
				if err != nil {
					return err
				}

				_, err = file.Write(rawData)
				if err != nil {
					return err
				}
				_, err = file.WriteString("\n")
				if err != nil {
					return err
				}

				log.Println(string(rawData))

				if statusResp != nil {
					lastStatusResp = statusResp
				}
				if netInfo != nil {
					lastNetInfo = netInfo
				}
				select {
				case <-cmd.Context().Done():
					return nil
				case <-time.After(time.Second * 30):
					continue
				}
			}
		},
	}

	return command
}

type PeerData struct {
	PeerCount int       `json:"peer_count"`
	Height    int64     `json:"height"`
	LocalTime time.Time `json:"local_time"`
}
