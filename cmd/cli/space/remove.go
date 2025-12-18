package space

import (
	"fmt"

	"github.com/alanshaw/buff/pkg/fx/cli"
	dlgstore "github.com/alanshaw/buff/pkg/store/delegation"
	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/principal"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove <space-did>",
	Aliases: []string{"rm"},
	Short:   "List known spaces",
	Args:    cobra.ExactArgs(1),
	RunE:    cli.FXCommand(doRemove),
}

func doRemove(cmd *cobra.Command, args []string, id principal.Signer, delegationStore dlgstore.Store) error {
	space, err := did.Parse(args[0])
	cobra.CheckErr(err)

	n := 0
	for dlg, err := range delegationStore.List(cmd.Context(), id) {
		cobra.CheckErr(err)
		if dlg.Subject() != nil && dlg.Subject().DID() == space.DID() {
			err := delegationStore.Del(cmd.Context(), dlg.Link())
			cobra.CheckErr(err)
			n++
		}
	}

	switch n {
	case 0:
		return fmt.Errorf("no delegation found for space: %s", space)
	case 1:
		cmd.Println("Removed 1 delegation")
	default:
		cmd.Printf("Removed %d delegations\n", n)
	}
	return nil
}
