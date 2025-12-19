package upload

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/alanshaw/buff/pkg/config/app"
	"github.com/alanshaw/buff/pkg/fx/cli"
	rcpt_client "github.com/alanshaw/buff/pkg/receipt"
	dstore "github.com/alanshaw/buff/pkg/store/delegation"
	assert_caps "github.com/alanshaw/libracha/capabilities/assert"
	"github.com/alanshaw/libracha/capabilities/blob"
	http_caps "github.com/alanshaw/libracha/capabilities/http"
	"github.com/alanshaw/libracha/digestutil"
	ucanlib "github.com/alanshaw/libracha/ucan"
	"github.com/alanshaw/ucantone/client"
	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/execution"
	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/principal"
	"github.com/alanshaw/ucantone/principal/ed25519"
	"github.com/alanshaw/ucantone/result"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/delegation"
	"github.com/alanshaw/ucantone/ucan/delegation/policy"
	"github.com/alanshaw/ucantone/ucan/invocation"
	"github.com/alanshaw/ucantone/ucan/receipt"
	"github.com/ipfs/go-cid"
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

func doUpload(cmd *cobra.Command, args []string, id principal.Signer, serviceConfig app.ExternalServicesConfig, delegationStore dstore.Store) error {
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
		invocation.WithAudience(serviceConfig.Upload.ID),
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
		serviceConfig.Upload.ID,
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
		serviceConfig.Upload.ID,
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

	client, err := client.NewHTTP(serviceConfig.Upload.URL)
	cobra.CheckErr(err)

	request := execution.NewRequest(
		cmd.Context(),
		inv,
		execution.WithProofs(proofs...),
		execution.WithDelegations(delegations...),
	)

	response, err := client.Execute(request)
	cobra.CheckErr(err)

	addOut, err := result.MapResultR1(
		response.Out(),
		func(o ipld.Any) (blob.AddOK, error) {
			model := blob.AddOK{}
			err = datamodel.Rebind(datamodel.NewAny(o), &model)
			return model, err
		},
		func(x ipld.Any) (error, error) {
			return nil, fmt.Errorf("failed %q task: %+v", blob.AddCommand, x)
		},
	)
	cobra.CheckErr(err)
	addOK, _ := result.Unwrap(addOut)

	// Find /http/put invocation
	var httpPutInv ucan.Invocation
	for _, inv := range response.Metadata().Invocations() {
		if inv.Command() == http_caps.PutCommand {
			httpPutInv = inv
			break
		}
	}
	if httpPutInv == nil {
		return fmt.Errorf("missing %q invocation in response", http_caps.PutCommand)
	}
	blobProvider, err := extractBlobProviderKey(httpPutInv)
	if err != nil {
		return fmt.Errorf("extracting blob provider key: %w", err)
	}

	// Find allocation receipt in the response metadata
	var allocRcpt ucan.Receipt
	for _, inv := range response.Metadata().Invocations() {
		if inv.Command() != blob.AllocateCommand {
			continue
		}
		rcpt, ok := response.Metadata().Receipt(inv.Task().Link())
		if ok {
			allocRcpt = rcpt
			break
		}
	}
	if allocRcpt == nil {
		return fmt.Errorf("missing %q receipt in response", blob.AllocateCommand)
	}

	allocOut, err := result.MapResultR1(
		allocRcpt.Out(),
		func(o ipld.Any) (blob.AllocateOK, error) {
			model := blob.AllocateOK{}
			err = datamodel.Rebind(datamodel.NewAny(o), &model)
			return model, err
		},
		func(x ipld.Any) (error, error) {
			return nil, fmt.Errorf("failed %q task: %+v", blob.AllocateCommand, x)
		},
	)
	cobra.CheckErr(err)
	allocOK, _ := result.Unwrap(allocOut)

	var httpPutRcpt ucan.Receipt
	if allocOK.Address == nil {
		cmd.Printf("✅ skipping upload, %q already has %q.\n", allocRcpt.Issuer().DID(), digestutil.Format(digest))
	} else {
		cmd.Printf("⬆️ uploading %q to %q (%s)\n", digestutil.Format(digest), allocRcpt.Issuer().DID(), allocOK.Address.URL.URL().String())
		putReq, err := http.NewRequestWithContext(cmd.Context(), http.MethodPut, allocOK.Address.URL.URL().String(), bytes.NewReader(data))
		cobra.CheckErr(err)
		for k, v := range allocOK.Address.Headers {
			putReq.Header.Set(k, v)
		}
		putRes, err := http.DefaultClient.Do(putReq)
		cobra.CheckErr(err)
		defer putRes.Body.Close()

		if putRes.StatusCode < 200 || putRes.StatusCode >= 300 {
			body, _ := io.ReadAll(putRes.Body)
			return fmt.Errorf("upload failed with status %d: %s", putRes.StatusCode, string(body))
		}

		cmd.Printf("🧾 issuing receipt for completed %q task", http_caps.PutCommand)
		httpPutRcpt, err = receipt.Issue(
			blobProvider,
			httpPutInv.Task().Link(),
			result.OK[ipld.Map, ipld.Any](ipld.Map{}),
		)
		cobra.CheckErr(err)

		request := execution.NewRequest(cmd.Context(), httpPutRcpt)
		_, err = client.Execute(request)
		cobra.CheckErr(err)
	}

	cmd.Printf("⏳ awaiting site from %q task: %s\n", blob.AcceptCommand, addOK.Site.Task)

	rcptClient := rcpt_client.New(serviceConfig.Upload.URL.JoinPath("receipt"))
	accRcpt, accRcptCt, err := rcptClient.Poll(cmd.Context(), addOK.Site.Task)
	cobra.CheckErr(err)

	accOut, err := result.MapResultR1(
		accRcpt.Out(),
		func(o ipld.Any) (blob.AcceptOK, error) {
			model := blob.AcceptOK{}
			err = datamodel.Rebind(datamodel.NewAny(o), &model)
			return model, err
		},
		func(x ipld.Any) (error, error) {
			return nil, fmt.Errorf("failed %q task: %+v", blob.AcceptCommand, x)
		},
	)
	cobra.CheckErr(err)
	accOK, _ := result.Unwrap(accOut)

	cmd.Printf("✍️ location commitment: %s\n", accOK.Site)

	var locationCommitment ucan.Invocation
	for _, inv := range accRcptCt.Invocations() {
		if inv.Link() == accOK.Site {
			locationCommitment = inv
		}
	}
	if locationCommitment == nil {
		return fmt.Errorf("missing location commitment\n")
	}

	loc := assert_caps.LocationArguments{}
	err = datamodel.Rebind(datamodel.NewAny(locationCommitment.Arguments()), &loc)
	cobra.CheckErr(err)

	for _, location := range loc.Location {
		cmd.Printf("📍 blob location: %s\n", location.URL().String())
	}

	cmd.Printf("✅ upload complete! Blob %q accepted in space %q\n", digestutil.Format(digest), space)
	cmd.Printf("🌱 %s\n", cid.NewCidV1(cid.Raw, digest))

	return nil
}

// extractBlobProviderKey extracts the blob provider's signing key from the
// /http/put invocation metadata.
func extractBlobProviderKey(inv ucan.Invocation) (principal.Signer, error) {
	if _, ok := inv.Metadata()["keys"]; !ok {
		return nil, fmt.Errorf("missing 'keys' metadata")
	}
	keyMap, ok := inv.Metadata()["keys"].(ipld.Map)
	if !ok {
		return nil, fmt.Errorf("invalid 'keys' metadata: not an IPLD map")
	}
	val, ok := keyMap[inv.Issuer().DID().String()]
	if !ok {
		return nil, fmt.Errorf("missing private key for %q in 'keys' metadata", inv.Issuer().DID().String())
	}
	keyBytes, ok := val.([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid private key for %q in 'keys' metadata: not a byte slice", inv.Issuer().DID().String())
	}
	return ed25519.Decode(keyBytes)
}
