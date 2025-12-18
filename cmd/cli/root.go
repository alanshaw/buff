package cli

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"strings"

	logging "github.com/ipfs/go-log/v2"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/alanshaw/buff/cmd/cli/space"
	"github.com/alanshaw/buff/cmd/cli/upload"
	"github.com/alanshaw/buff/pkg/build"
)

func ExecuteContext(ctx context.Context) {
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.Fatal(err)
	}
}

var log = logging.Logger("cmd")

var configFilePath = path.Join("buff", "config.toml")

var (
	cfgFile  string
	logLevel string
	rootCmd  = &cobra.Command{
		Use:     "buff",
		Short:   "A client for the Storacha Network",
		Long:    "UCAN 1.0 compatible client for the Storacha Network",
		Version: build.Version,
	}
)

func init() {
	cobra.OnInitialize(initLogging, initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file path. Attempts to load from user config directory if not set e.g. ~/.config/"+configFilePath)

	rootCmd.PersistentFlags().String("data-dir", filepath.Join(lo.Must(os.UserHomeDir()), ".buff"), "Upload service data directory")
	cobra.CheckErr(viper.BindPFlag("repo.data_dir", rootCmd.PersistentFlags().Lookup("data-dir")))

	rootCmd.PersistentFlags().String("key-file", "", "Path to a PEM file containing ed25519 private key")
	cobra.CheckErr(rootCmd.MarkPersistentFlagFilename("key-file", "pem"))
	cobra.CheckErr(viper.BindPFlag("identity.key_file", rootCmd.PersistentFlags().Lookup("key-file")))

	// register all commands and their subcommands
	rootCmd.AddCommand(space.Cmd)
	rootCmd.AddCommand(upload.Cmd)
}

func initConfig() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("BUFF")

	if cfgFile == "" {
		if configDir, err := os.UserConfigDir(); err == nil {
			defaultCfgFile := path.Join(configDir, configFilePath)
			if inf, err := os.Stat(defaultCfgFile); err == nil && !inf.IsDir() {
				log.Infof("loading config automatically from: %s", defaultCfgFile)
				cfgFile = defaultCfgFile
			}
		}
	}

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		cobra.CheckErr(viper.ReadInConfig())
	} else {
		// otherwise look for buff-config.toml in current directory
		viper.SetConfigName("buff-config")
		viper.SetConfigType("toml")
		viper.AddConfigPath(".")
		// Don't error if config file is not found - it's optional
		_ = viper.ReadInConfig()
	}
}

func initLogging() {
	if logLevel != "" {
		ll, err := logging.LevelFromString(logLevel)
		cobra.CheckErr(err)
		logging.SetAllLoggers(ll)
	} else {
		// else set all loggers to warn level, then the ones we care most about to info.
		logging.SetAllLoggers(logging.LevelWarn)
		logging.SetLogLevel("cmd/upload", "info")
		logging.SetLogLevel("pkg/store/delegation", "info")
	}
}
