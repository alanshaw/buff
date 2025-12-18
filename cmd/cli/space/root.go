package space

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "space",
	Short: "Manage spaces",
}

func init() {
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(removeCmd)
}
