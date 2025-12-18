package delegation

import (
	"context"
	"iter"

	ucanlib "github.com/alanshaw/libracha/ucan"
	"github.com/alanshaw/ucantone/ucan"
)

const NullSubject = "null"

type Store interface {
	ucanlib.DelegationFinder
	Del(ctx context.Context, root ucan.Link) error
	Get(ctx context.Context, root ucan.Link) (ucan.Delegation, error)
	Put(ctx context.Context, dlg ucan.Delegation) error
	List(ctx context.Context, aud ucan.Principal) iter.Seq2[ucan.Delegation, error]
}
