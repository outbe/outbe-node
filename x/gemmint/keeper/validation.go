package keeper

import (
	"errors"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/outbe/outbe-node/x/gemmint/constants"
)

func (k Keeper) ValidateContractAddress(addr string) error {

	if !strings.HasPrefix(addr, constants.ExpectedPrefix) {
		return errors.New("address must start with prefix 'outbe'")
	}

	if len(addr) != constants.ExpectedLength {
		return errors.New("invalid address length")
	}

	return nil
}

func (k Keeper) IsValidCreator(creator string) bool {
	_, err := sdk.AccAddressFromBech32(creator)
	return err == nil
}
