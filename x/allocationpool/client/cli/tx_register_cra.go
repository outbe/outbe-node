package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/outbe/outbe-node/x/allocationpool/types"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdRegisterCRA() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-cra [cra_address]",
		Short: "Register a cra",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgRegisterCRA{
				CraAddress: args[0],
			}

			// Broadcast the transaction
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	// Add transaction flags
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
