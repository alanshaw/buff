package app

import (
	"net/url"

	"github.com/alanshaw/ucantone/did"
)

type ExternalServicesConfig struct {
	Indexer IndexingServiceConfig
	Upload  UploadServiceConfig
}

type IndexingServiceConfig struct {
	ID  did.DID
	URL *url.URL
}

type UploadServiceConfig struct {
	ID  did.DID
	URL *url.URL
}
