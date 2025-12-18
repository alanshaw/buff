package app

import (
	"github.com/alanshaw/ucantone/principal"
)

// IdentityConfig contains identity-related configuration
type IdentityConfig struct {
	Signer principal.Signer
}
