package identity

import (
	"github.com/alanshaw/buff/pkg/config/app"
	"github.com/alanshaw/ucantone/principal"
	"go.uber.org/fx"
)

var Module = fx.Module("identity",
	fx.Provide(ProvideIdentity),
)

func ProvideIdentity(cfg app.IdentityConfig) principal.Signer {
	return cfg.Signer
}
