package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/outbe/outbe-node/x/cra/types"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdWalletReward() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wallet_reward [address]",
		Short: "claim wallet reward",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgWalletReward{
				Creator: clientCtx.GetFromAddress().String(),
				Address: args[0],
			}

			// Broadcast the transaction
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	// Add transaction flags
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
