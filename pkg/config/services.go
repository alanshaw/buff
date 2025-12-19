package config

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/alanshaw/ucantone/did"

	"github.com/alanshaw/buff/pkg/config/app"
)

type ServicesConfig struct {
	Indexer IndexingServiceConfig `mapstructure:"indexer" validate:"required" toml:"indexer,omitempty"`
	Upload  UploadServiceConfig   `mapstructure:"upload" validate:"required" toml:"upload,omitempty"`
}

func (s ServicesConfig) Validate() error {
	return validateConfig(s)
}

func (s ServicesConfig) ToAppConfig() (app.ExternalServicesConfig, error) {
	var (
		out app.ExternalServicesConfig
		err error
	)

	out.Upload, err = s.Upload.ToAppConfig()
	if err != nil {
		return app.ExternalServicesConfig{}, fmt.Errorf("creating upload service app config: %w", err)
	}
	out.Indexer, err = s.Indexer.ToAppConfig()
	if err != nil {
		return app.ExternalServicesConfig{}, fmt.Errorf("creating indexing service app config: %w", err)
	}

	return out, nil
}

type IndexingServiceConfig struct {
	ID  string `mapstructure:"id" validate:"required" flag:"indexing-service-id" toml:"id,omitempty"`
	URL string `mapstructure:"url" validate:"required,url" flag:"indexing-service-url" toml:"url,omitempty"`
}

func (s *IndexingServiceConfig) Validate() error {
	return validateConfig(s)
}

func (s *IndexingServiceConfig) ToAppConfig() (app.IndexingServiceConfig, error) {
	sid, err := did.Parse(s.ID)
	if err != nil {
		return app.IndexingServiceConfig{}, fmt.Errorf("parsing indexing service DID: %w", err)
	}
	var surl *url.URL
	if s.URL == "" {
		if !strings.HasPrefix(s.ID, "did:web:") {
			return app.IndexingServiceConfig{}, fmt.Errorf("indexing service URL is required for non-web DIDs")
		}
		surl, err = url.Parse(fmt.Sprintf("https://%s", strings.TrimPrefix(s.ID, "did:web:")))
		if err != nil {
			return app.IndexingServiceConfig{}, fmt.Errorf("parsing indexing service URL from DID: %w", err)
		}
	} else {
		surl, err = url.Parse(s.URL)
		if err != nil {
			return app.IndexingServiceConfig{}, fmt.Errorf("parsing indexing service URL: %w", err)
		}
	}
	return app.IndexingServiceConfig{ID: sid, URL: surl}, nil
}

type UploadServiceConfig struct {
	ID  string `mapstructure:"id" validate:"required" flag:"upload-service-id" toml:"id,omitempty"`
	URL string `mapstructure:"url" validate:"required,url" flag:"upload-service-url" toml:"url,omitempty"`
}

func (s *UploadServiceConfig) Validate() error {
	return validateConfig(s)
}

func (s *UploadServiceConfig) ToAppConfig() (app.UploadServiceConfig, error) {
	sdid, err := did.Parse(s.ID)
	if err != nil {
		return app.UploadServiceConfig{}, fmt.Errorf("parsing upload service DID: %w", err)
	}
	var surl *url.URL
	if s.URL == "" {
		if !strings.HasPrefix(s.ID, "did:web:") {
			return app.UploadServiceConfig{}, fmt.Errorf("upload service URL is required for non-web DIDs")
		}
		surl, err = url.Parse(fmt.Sprintf("https://%s", strings.TrimPrefix(s.ID, "did:web:")))
		if err != nil {
			return app.UploadServiceConfig{}, fmt.Errorf("parsing upload service URL from DID: %w", err)
		}
	} else {
		surl, err = url.Parse(s.URL)
		if err != nil {
			return app.UploadServiceConfig{}, fmt.Errorf("parsing upload service URL: %w", err)
		}
	}
	return app.UploadServiceConfig{ID: sdid, URL: surl}, nil
}
