package slashing

import (
	"bytes"
	"context"

	"cosmossdk.io/core/comet"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// BeginBlocker check for infraction evidence or downtime of validators
// on every begin block
func BeginBlocker(ctx context.Context, k keeper.Keeper) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, telemetry.Now(), telemetry.MetricKeyBeginBlocker)

	stakingVoteInfo, err := k.GetStakingVoteInfo(ctx)
	if err != nil {
		return err
	}

	// Iterate over all the validators which *should* have signed this block
	// store whether or not they have actually signed it and slash/unbond any
	// which have missed too many blocks in a row (downtime slashing)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	for _, voteInfo := range stakingVoteInfo {
		var blockIdFlag comet.BlockIDFlag
		for _, vote := range sdkCtx.VoteInfos() {
			if bytes.Equal(vote.Validator.Address, voteInfo.Validator.Address) {
				blockIdFlag = comet.BlockIDFlag(vote.BlockIdFlag)
				break
			}
		}

		if err := k.HandleValidatorSignature(ctx, voteInfo.Validator.Address, voteInfo.Validator.Power, blockIdFlag); err != nil {
			return err
		}
	}

	return nil
}
