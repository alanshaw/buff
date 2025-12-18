package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alanshaw/buff/pkg/store/delegation"
	leveldb "github.com/ipfs/go-ds-leveldb"
	"go.uber.org/fx"

	"github.com/alanshaw/buff/pkg/config/app"
)

var Module = fx.Module("store",
	fx.Provide(
		ProvideConfigs,
		NewDelegationStore,
	),
)

type Configs struct {
	fx.Out
	Delegation app.DelegationStorageConfig
}

// ProvideConfigs provides the fields of a storage config
func ProvideConfigs(cfg app.StorageConfig) Configs {
	return Configs{
		Delegation: cfg.Delegation,
	}
}

func NewDelegationStore(cfg app.DelegationStorageConfig, lc fx.Lifecycle) (delegation.Store, error) {
	if cfg.Dir == "" {
		return nil, fmt.Errorf("no data dir provided for provider store")
	}

	ds, err := newDatastore(cfg.Dir)
	if err != nil {
		return nil, fmt.Errorf("creating provider store: %w", err)
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return ds.Close()
		},
	})

	return delegation.NewDSDelegationStore(ds), nil
}

func newDatastore(path string) (*leveldb.Datastore, error) {
	dirPath, err := mkdirp(path)
	if err != nil {
		return nil, fmt.Errorf("creating leveldb for store at path %s: %w", path, err)
	}
	return leveldb.NewDatastore(dirPath, nil)
}

func mkdirp(dirpath ...string) (string, error) {
	dir := filepath.Join(dirpath...)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return "", fmt.Errorf("creating directory: %s: %w", dir, err)
	}
	return dir, nil
}
