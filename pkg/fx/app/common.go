package app

import (
	"github.com/alanshaw/buff/pkg/config/app"
	"github.com/alanshaw/buff/pkg/fx/identity"
	"github.com/alanshaw/buff/pkg/fx/store"
	"go.uber.org/fx"
)

func CommonModules(cfg app.AppConfig) fx.Option {
	return fx.Module("common",
		// Supply top level config, and it's sub-configs
		// this allows dependencies to be taken on, for example, app.IdentityConfig or app.ServerConfig
		// instead of needing to depend on the top level app.AppConfig
		fx.Supply(cfg),
		fx.Supply(cfg.Identity),
		fx.Supply(cfg.Storage),

		identity.Module,
		store.Module,
	)
}
