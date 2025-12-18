package space

import (
	"github.com/alanshaw/buff/pkg/fx/cli"
	dlgstore "github.com/alanshaw/buff/pkg/store/delegation"
	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/principal"
	"github.com/alanshaw/ucantone/principal/ed25519"
	"github.com/alanshaw/ucantone/ucan/command"
	"github.com/alanshaw/ucantone/ucan/delegation"
	"github.com/algorand/go-algorand-sdk/mnemonic"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new space",
	Args:  cobra.MaximumNArgs(1),
	RunE:  cli.FXCommand(doCreate),
}

func doCreate(cmd *cobra.Command, args []string, id principal.Signer, delegationStore dlgstore.Store) error {
	signer, err := ed25519.Generate()
	cobra.CheckErr(err)

	name := ""
	if len(args) > 0 {
		name = args[0]
	}

	dlg, err := delegation.Delegate(
		signer,
		id.DID(),
		signer,
		command.Top(),
		delegation.WithMetadata(ipld.Map{"name": name}),
		delegation.WithNoExpiration(),
	)
	cobra.CheckErr(err)

	err = delegationStore.Put(cmd.Context(), dlg)
	cobra.CheckErr(err)

	cmd.Println("Space ID:")
	cmd.Println(signer.DID())
	cmd.Println("")
	cmd.Println("Recovery phrase:")
	mnemonic, err := mnemonic.FromKey(signer.Raw())
	cobra.CheckErr(err)
	cmd.Println(mnemonic)

	return nil
}
