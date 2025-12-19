package config

import (
	"github.com/alanshaw/buff/pkg/presets"
	"github.com/spf13/viper"
)

func LoadPresets() error {
	networkStr := viper.GetString("network")
	network, err := presets.ParseNetwork(networkStr)
	if err != nil {
		return err
	}

	preset, err := presets.GetPreset(network)
	if err != nil {
		return err
	}

	viper.SetDefault("services.indexer.id", preset.Services.IndexingServiceID.String())
	viper.SetDefault("services.indexer.url", preset.Services.IndexingServiceURL.String())

	viper.SetDefault("services.upload.id", preset.Services.UploadServiceID.String())
	viper.SetDefault("services.upload.url", preset.Services.UploadServiceURL.String())

	return nil
}
