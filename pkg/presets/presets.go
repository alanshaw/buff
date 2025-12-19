package presets

import (
	"fmt"
	"net/url"

	"github.com/alanshaw/ucantone/did"
	"github.com/samber/lo"
)

// Network represents the network the node will operate on
type Network string

const (
	Dev Network = "dev"
)

var AvailableNetworks = []Network{Dev}

// String returns the string representation of the network
func (n Network) String() string {
	switch n {
	case Dev:
		return string(n)
	default:
		return "unknown"
	}
}

// ParseNetwork parses a string into a Network type
func ParseNetwork(s string) (Network, error) {
	switch s {
	case string(Dev):
		return Dev, nil
	default:
		return Network(""), fmt.Errorf("unknown network: %q (valid networks are: %q)", s, AvailableNetworks)
	}
}

// ServiceSettings holds the service configuration for a network
type ServiceSettings struct {
	IndexingServiceID  did.DID
	IndexingServiceURL *url.URL
	UploadServiceID    did.DID
	UploadServiceURL   *url.URL
}

// Preset holds all configuration presets for a network
type Preset struct {
	Services ServiceSettings
}

// Development service preset values (default)
func devServiceSettings() ServiceSettings {
	indexingServiceID := lo.Must(did.Parse("did:web:indexer.dev.storacha.network"))
	indexingServiceURL := lo.Must(url.Parse("http://indexer.dev.storacha.network"))

	uploadServiceID := lo.Must(did.Parse("did:web:up.dev.storacha.network"))
	uploadServiceURL := lo.Must(url.Parse("http://up.dev.storacha.network"))

	return ServiceSettings{
		IndexingServiceID:  indexingServiceID,
		IndexingServiceURL: indexingServiceURL,
		UploadServiceID:    uploadServiceID,
		UploadServiceURL:   uploadServiceURL,
	}
}

// GetPreset returns the complete preset configuration for a given network
func GetPreset(network Network) (Preset, error) {
	switch network {
	case Dev:
		return Preset{
			Services: devServiceSettings(),
		}, nil
	default:
		return Preset{}, fmt.Errorf("unknown network: %s", network)
	}
}
