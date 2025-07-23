package wasmbinding

import (
	"encoding/json"
	"log"
	"strconv"

	errortypes "github.com/outbe/outbe-node/errors"
	"github.com/outbe/outbe-node/wasmbinding/bindings"
	randKeepers "github.com/outbe/outbe-node/x/rand/keeper"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func CustomQuerier(qp *QueryPlugin) func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
	return func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
		var contractQuery bindings.OutbeQuery
		if err := json.Unmarshal(request, &contractQuery); err != nil {
			return nil, sdkerrors.Wrapf(errortypes.ErrInvalidType, "[CustomQuerier][Unmarshal Contract Query Result] failed. Contract query is not valid, couldn't be parsed.")
		}

		switch {
		case contractQuery.Randomness != nil:

			randomness, err := GetRandomness(ctx, *qp.keeper)
			if err != nil {
				return nil, sdkerrors.Wrap(errortypes.ErrInvalidType, "[CustomQuerier][GetMinter] failed.")
			}

			response := bindings.RandomnessResponse{Period: randomness.Period, Randomness: randomness.Randomness}

			bz, err := json.Marshal(response)
			if err != nil {
				return nil, sdkerrors.Wrapf(errortypes.ErrInvalidType, "[CustomQuerier][Marshal] failed to marshal response")
			}

			return bz, nil

		default:
			return nil, sdkerrors.Wrapf(errortypes.ErrInvalidType, "[CustomQuerier] failed. unknown gemchain query variante.")
		}
	}
}

func GetRandomness(ctx sdk.Context, randKeeper randKeepers.Keeper) (bindings.RandomnessResponse, error) {

	log.Println("############## Smart contract query for fetching randomness is Started ##############")

	var response bindings.RandomnessResponse

	logger := randKeeper.Logger(ctx)

	randomness, _ := randKeeper.GetPeriod(ctx)

	if logger != nil {
		logger.Info("Fetching smart contract query for minter successfully done.",
			"query", "GetAllMinter",
		)
	}

	priod := strconv.FormatUint(randomness.CurrentPeriod, 10)
	response.Period = priod
	response.Randomness = string(randomness.CurrentSeed)

	log.Println("############## End of Smart contract query for fetching randomness ##############")

	return response, nil
}
