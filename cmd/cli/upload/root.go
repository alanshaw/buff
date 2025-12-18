package upload

import (
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/alanshaw/buff/pkg/fx/cli"
	dstore "github.com/alanshaw/buff/pkg/store/delegation"
	"github.com/alanshaw/libracha/capabilities/blob"
	ucanlib "github.com/alanshaw/libracha/ucan"
	"github.com/alanshaw/ucantone/client"
	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/execution"
	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/principal"
	"github.com/alanshaw/ucantone/result"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/delegation"
	"github.com/alanshaw/ucantone/ucan/delegation/policy"
	"github.com/alanshaw/ucantone/ucan/invocation"
	logging "github.com/ipfs/go-log/v2"
	"github.com/multiformats/go-multihash"
	"github.com/spf13/cobra"
)

var log = logging.Logger("cmd/upload")

var Cmd = &cobra.Command{
	Use:     "upload <space-did> [<file-path>]",
	Aliases: []string{"up"},
	Short:   "Upload files to the Storacha Network",
	Args:    cobra.MinimumNArgs(1),
	RunE:    cli.FXCommand(doUpload),
}

func doUpload(cmd *cobra.Command, args []string, id principal.Signer, delegationStore dstore.Store) error {
	space, err := did.Parse(args[0])
	cobra.CheckErr(err)

	var data []byte
	if len(args) == 1 {
		b, err := io.ReadAll(cmd.InOrStdin())
		cobra.CheckErr(err)
		data = b
	} else {
		b, err := os.ReadFile(args[1])
		cobra.CheckErr(err)
		data = b
	}

	digest, err := multihash.Sum(data, multihash.SHA2_256, -1)
	cobra.CheckErr(err)

	serviceID, err := did.Parse("did:key:z6MkiZfWmWbXpBj2bxF4w8ifBRi8PRSa83qUFTWq7rb73Hse")
	cobra.CheckErr(err)

	serviceURL, err := url.Parse("http://localhost:3000")
	cobra.CheckErr(err)

	matcher := ucanlib.NewDelegationMatcher(delegationStore)
	proofs, proofLinks, err := ucanlib.ProofChain(cmd.Context(), matcher, id, blob.AddCommand, space)
	cobra.CheckErr(err)

	if len(proofs) == 0 {
		return fmt.Errorf("missing %q delegations for space: %s", blob.AddCommand, space)
	}

	inv, err := blob.Add.Invoke(
		id,
		space,
		&blob.AddArguments{
			Blob: blob.Blob{
				Digest: digest,
				Size:   uint64(len(data)),
			},
		},
		invocation.WithAudience(serviceID),
		invocation.WithProofs(proofLinks...),
	)
	cobra.CheckErr(err)

	// create required invocation delegations
	// TODO: get proof chain for these as well and add to invocation - for now it
	// is fine as we know we have top authority over the space so this delegation
	// will be included already.
	delegations := []ucan.Delegation{}
	dlg, err := delegation.Delegate(
		id,
		serviceID,
		space,
		blob.AllocateCommand,
		delegation.WithPolicyBuilder(
			policy.And(
				policy.Equal(".blob.digest", []byte(digest)),
				policy.Equal(".blob.size", len(data)),
			),
		),
	)
	cobra.CheckErr(err)
	delegations = append(delegations, dlg)

	dlg, err = delegation.Delegate(
		id,
		serviceID,
		space,
		blob.AcceptCommand,
		delegation.WithPolicyBuilder(
			policy.And(
				policy.Equal(".blob.digest", []byte(digest)),
				policy.Equal(".blob.size", len(data)),
			),
		),
	)
	cobra.CheckErr(err)
	delegations = append(delegations, dlg)

	client, err := client.NewHTTP(serviceURL)
	cobra.CheckErr(err)

	request := execution.NewRequest(
		cmd.Context(),
		inv,
		execution.WithProofs(proofs...),
		execution.WithDelegations(delegations...),
	)

	response, err := client.Execute(request)
	cobra.CheckErr(err)

	result.MatchResultR0(
		response.Result(),
		func(o ipld.Any) {
			args := blob.AddOK{}
			err := datamodel.Rebind(datamodel.NewAny(o), &args)
			cobra.CheckErr(err)
			cmd.Println("Blob add invocation successful", args)
		},
		func(x ipld.Any) {
			cmd.Printf("Invocation failed: %+v\n", x)
		},
	)

	return nil
}
