package config

import (
	"github.com/alanshaw/buff/pkg/config/app"
	"github.com/alanshaw/buff/pkg/config/lib"
)

type IdentityConfig struct {
	KeyFile string `mapstructure:"key_file" validate:"required" flag:"key-file" toml:"key_file"`
}

func (i IdentityConfig) Validate() error {
	return validateConfig(i)
}

func (i IdentityConfig) ToAppConfig() (app.IdentityConfig, error) {
	id, err := lib.SignerFromEd25519PEMFile(i.KeyFile)
	if err != nil {
		return app.IdentityConfig{}, err
	}
	return app.IdentityConfig{
		Signer: id,
	}, nil
}
