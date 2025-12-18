package space

import (
	"fmt"

	"github.com/alanshaw/buff/pkg/fx/cli"
	dlgstore "github.com/alanshaw/buff/pkg/store/delegation"
	"github.com/alanshaw/ucantone/principal"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List known spaces",
	Args:    cobra.NoArgs,
	RunE:    cli.FXCommand(doList),
}

func doList(cmd *cobra.Command, id principal.Signer, delegationStore dlgstore.Store) error {
	for dlg, err := range delegationStore.List(cmd.Context(), id) {
		cobra.CheckErr(err)
		if dlg.Metadata() != nil {
			v, ok := dlg.Metadata()["name"]
			if ok {
				if str, ok := v.(string); ok {
					cmd.Println(fmt.Sprintf("%s %s", dlg.Subject(), str))
					continue
				}
			}
		}
		cmd.Println(dlg.Subject())
	}
	return nil
}
