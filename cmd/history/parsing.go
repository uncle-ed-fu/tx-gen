package history

import (
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/montanaflynn/stats"
	coretypes "github.com/tendermint/tendermint/types"
)

type BlockSummary struct {
	Height   int64  `json:"height"`
	Proposer string `json:"proposer"`
	// Signers depicts if a validator signed the last block
	Signers            map[string]bool `json:"signers"`
	MissedSigners      int64           `json:"missed_signers"`
	Time               time.Time       `json:"time"`
	TimeSinceLastBlock float64         `json:"tslb"`
}

type ChainSummary struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`

	// in seconds
	AverageBlockTime  float64 `json:"average_block_time"`
	LongestBlockTime  float64 `json:"longest_block_time"`
	ShortestBlockTime float64 `json:"shortest_block_time"`

	MedianBlockTime            float64 `json:"median_block_time"`
	StandardDeviationBlockTime float64 `json:"std_dev"`

	AverageMissedSigners float64         `json:"average_missed_signers"`
	ProposerRepeats      int             `json:"proposer_repeats"`
	SignerTally          map[string]uint // not including these as json atm
	ProposerTally        map[string]uint
}

func parseSigners(valNames map[string]string, block *coretypes.Block) (map[string]bool, int64) {
	out := make(map[string]bool)
	missedSigners := int64(0)
	for _, sig := range block.LastCommit.Signatures {
		if sig.Absent() {
			missedSigners++
			continue
		}
		valAddr, err := sdk.ConsAddressFromHex(sig.ValidatorAddress.String())
		if err != nil {
			panic(err)
		}

		valName, has := valNames[valAddr.String()]
		if !has {
			valName = valAddr.String()
		}

		switch sig.BlockIDFlag {
		case coretypes.BlockIDFlagCommit:
			out[valName] = true
		case coretypes.BlockIDFlagAbsent:
			out[valName] = false
		case coretypes.BlockIDFlagNil:
			out[valName] = false
			missedSigners++
		default:
			panic("unknown block id flag")
		}
	}

	return out, missedSigners
}

func SummarizeBlocks(blocks ...BlockSummary) (ChainSummary, []BlockSummary) {
	if len(blocks) == 0 {
		return ChainSummary{}, blocks
	}

	times := make([]int64, len(blocks)-1)
	missedSigners := make([]int64, len(blocks))

	// signing and block creations
	signerTally := make(map[string]uint)
	proposerTally := make(map[string]uint)

	proposerRepeats := 0

	lastProposer := ""
	secondLastProposer := ""
	for i, b := range blocks {
		// get the time difference
		if i != 0 {
			b1 := blocks[i-1].Time.UnixMilli()
			b2 := b.Time.UnixMilli()
			times[i-1] = b2 - b1
			blocks[i].TimeSinceLastBlock = float64(b2-b1) / 1000
		}
		for signer, signed := range b.Signers {
			if signed {
				signerTally[signer]++
			}
		}
		missedSigners[i] = b.MissedSigners
		proposerTally[b.Proposer]++

		// count the number of times we have the same proposer 3 times in a row
		if b.Proposer == lastProposer && lastProposer == secondLastProposer {
			proposerRepeats++
		}

		secondLastProposer = lastProposer
		lastProposer = b.Proposer
	}

	return ChainSummary{
		Start:                      blocks[0].Height,
		End:                        blocks[len(blocks)-1].Height,
		AverageBlockTime:           average(times),
		LongestBlockTime:           longest(times),
		ShortestBlockTime:          shortest(times),
		MedianBlockTime:            median(times),
		StandardDeviationBlockTime: stdDev(times),
		AverageMissedSigners:       average(missedSigners),
		SignerTally:                signerTally,
		ProposerTally:              proposerTally,
		ProposerRepeats:            proposerRepeats,
	}, blocks
}

// time in []milliseconds -> average time in seconds
func average(times []int64) float64 {
	total := int64(0)
	for _, v := range times {
		total += v
	}
	return (float64(total) / 1000) / float64(len(times))
}

// time in []milliseconds -> average time in seconds
func longest(times []int64) float64 {
	largest := int64(0)
	for _, v := range times {
		if v > largest {
			largest = v
		}
	}
	return (float64(largest) / 1000)
}

// time in []milliseconds -> average time in seconds
func shortest(times []int64) float64 {
	shortest := int64(100000000000000000)
	for _, v := range times {
		if v < shortest {
			shortest = v
		}
	}
	return (float64(shortest) / 1000)
}

func median(data []int64) float64 {
	dataCopy := make([]int64, len(data))
	copy(dataCopy, data)

	sort.Slice(dataCopy, func(i, j int) bool { return i < j })

	var median int64
	l := len(dataCopy)
	if l == 0 {
		return 0
	} else if l%2 == 0 {
		median = (dataCopy[l/2-1] + dataCopy[l/2]) / 2
	} else {
		median = dataCopy[l/2]
	}

	return float64(median) / 1000
}

func stdDev(times []int64) float64 {
	fData := make([]float64, len(times))
	for i, d := range times {
		fData[i] = float64(d)
	}
	res, err := stats.StandardDeviation(fData)
	if err != nil {
		panic(err)
	}
	return res / 1000
}
