package config

import (
	"os"
	"path/filepath"

	"github.com/alanshaw/buff/pkg/config/app"
)

type RepoConfig struct {
	DataDir string `mapstructure:"data_dir" validate:"required" flag:"data-dir" toml:"data_dir"`
}

func (r RepoConfig) Validate() error {
	return validateConfig(r)
}

func (r RepoConfig) ToAppConfig() (app.StorageConfig, error) {
	if r.DataDir == "" {
		return app.StorageConfig{}, nil
	}

	if err := os.MkdirAll(r.DataDir, 0755); err != nil {
		return app.StorageConfig{}, err
	}

	out := app.StorageConfig{
		DataDir: r.DataDir,
		Delegation: app.DelegationStorageConfig{
			Dir: filepath.Join(r.DataDir, "delegation", "datastore"),
		},
	}

	return out, nil
}
