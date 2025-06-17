package wasmbinding

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	errortypes "github.com/outbe/outbe-node/errors"
	"github.com/outbe/outbe-node/wasmbinding/bindings"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/outbe/outbe-node/x/pool/constants"
	Poolkeeper "github.com/outbe/outbe-node/x/pool/keeper"
)

func CustomQuerier(qp *QueryPlugin) func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
	return func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
		var contractQuery bindings.QcoreQuery
		if err := json.Unmarshal(request, &contractQuery); err != nil {
			return nil, sdkerrors.Wrapf(errortypes.ErrInvalidType, "[CustomQuerier][Unmarshal Contract Query Result] failed. Contract query is not valid, couldn't be parsed.")
		}

		switch {
		case contractQuery.QueryBlockEmissionRequest != nil:
			blockNumber := contractQuery.QueryBlockEmissionRequest.BlockNumber
			blockEmission, err := GetBlockEmission(ctx, blockNumber, *qp.keeper)
			if err != nil {
				return nil, sdkerrors.Wrap(errortypes.ErrInvalidType, err.Error())
			}

			response := bindings.QueryBlockEmissionResponse{BlockEmission: blockEmission.BlockEmission}

			bz, err := json.Marshal(response)
			if err != nil {
				return nil, sdkerrors.Wrap(errortypes.ErrInvalidType, "[CustomQuerier][Marshal] failed. Motus couldn't be marshaled response")
			}

			return bz, nil

		default:
			return nil, sdkerrors.Wrapf(errortypes.ErrInvalidType, "[CustomQuerier][GetBlockEmission] failed. unknown qcore query variante.")
		}
	}
}

func GetBlockEmission(ctx sdk.Context, blockNumber string, poolKeeper Poolkeeper.Keeper) (res bindings.QueryBlockEmissionResponse, err error) {

	log.Println("############## Smart contract query for fetching block emission is Started ##############")

	var response bindings.QueryBlockEmissionResponse

	logger := poolKeeper.Logger(ctx)
	num, err := strconv.ParseInt(blockNumber, 10, 64)
	if err != nil {
		fmt.Println("Error converting string to int64:", err)
		return
	}

	if ctx.BlockHeight() < constants.TransitionBlockNumber {
		if num < 0 {
			return bindings.QueryBlockEmissionResponse{}, errors.New("blocknumber is 0")
		}

		blockEmission, err := poolKeeper.CalculateExponentialBlockEmission(num)
		if err != nil {
			return bindings.QueryBlockEmissionResponse{}, errors.New("[Binding][GetBlockEmission][CalculateExponentialBlockEmission] failed.CalculateExponentialTokens failed")
		}
		response.BlockEmission = blockEmission
		return response, nil
	}

	blockEmission, err := poolKeeper.CalculateFixedBlockEmission(ctx)
	response.BlockEmission = blockEmission
	if err != nil {
		return bindings.QueryBlockEmissionResponse{}, errors.New("[Binding][GetBlockEmission][CalculateFixedBlockEmission] failed.CalculateExponentialTokens failed")
	}

	if logger != nil {
		logger.Info("Fetching block emission query successfully done.", "query", "GetBlockEmission", "Block Number:", blockNumber)
	}

	log.Println("############## End of Smart contract query for fetching block emission ##############")

	return response, nil
}
