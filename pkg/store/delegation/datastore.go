package delegation

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"strings"

	"github.com/alanshaw/buff/pkg/store"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/delegation"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("pkg/store/delegation")

type DSDelegationStore struct {
	ds datastore.Datastore
}

func NewDSDelegationStore(dstore datastore.Datastore) *DSDelegationStore {
	return &DSDelegationStore{dstore}
}

func (d *DSDelegationStore) Del(ctx context.Context, root ucan.Link) error {
	dlg, err := d.Get(ctx, root)
	if err != nil {
		if errors.Is(err, datastore.ErrNotFound) {
			return store.ErrNotFound
		}
		return err
	}
	if err := d.ds.Delete(ctx, queryKey(dlg)); err != nil {
		return err
	}
	return d.ds.Delete(ctx, datastore.NewKey(root.String()))
}

func (d *DSDelegationStore) Get(ctx context.Context, root ucan.Link) (ucan.Delegation, error) {
	b, err := d.ds.Get(ctx, datastore.NewKey(root.String()))
	if err != nil {
		if errors.Is(err, datastore.ErrNotFound) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}
	return delegation.Decode(b)
}

func (d *DSDelegationStore) List(ctx context.Context, aud ucan.Principal) iter.Seq2[ucan.Delegation, error] {
	log := log.With("aud", aud.DID().String())
	log.Infof("listing delegations")
	return func(yield func(ucan.Delegation, error) bool) {
		pfx := fmt.Sprintf("%s/", aud.DID().String())
		results, err := d.ds.Query(ctx, query.Query{Prefix: datastore.NewKey(pfx).String()})
		if err != nil {
			yield(nil, fmt.Errorf("querying datastore: %w", err))
			return
		}
		for entry := range results.Next() {
			if entry.Error != nil {
				yield(nil, fmt.Errorf("iterating query results: %w", entry.Error))
				return
			}
			dlg, err := delegation.Decode(entry.Value)
			if err != nil {
				yield(nil, fmt.Errorf("decoding delegation: %w", err))
				return
			}
			if !yield(dlg, nil) {
				return
			}
		}
	}
}

func (d *DSDelegationStore) FindByAudienceCommandSubject(ctx context.Context, aud ucan.Principal, cmd ucan.Command, sub ucan.Subject) iter.Seq2[ucan.Delegation, error] {
	log := log.With("aud", aud.DID().String(), "cmd", cmd)
	if sub != nil {
		log = log.With("sub", sub.DID().String())
	}
	log.Infof("finding delegations")
	return func(yield func(ucan.Delegation, error) bool) {
		pfx := fmt.Sprintf("%s/%s/", aud.DID().String(), sanitizeCommand(cmd))
		if sub != nil {
			pfx = fmt.Sprintf("%s%s/", pfx, sub.DID().String())
		}

		results, err := d.ds.Query(ctx, query.Query{Prefix: datastore.NewKey(pfx).String()})
		if err != nil {
			yield(nil, fmt.Errorf("querying datastore: %w", err))
			return
		}
		for entry := range results.Next() {
			if entry.Error != nil {
				yield(nil, fmt.Errorf("iterating query results: %w", entry.Error))
				return
			}
			dlg, err := delegation.Decode(entry.Value)
			if err != nil {
				yield(nil, fmt.Errorf("decoding delegation: %w", err))
				return
			}
			if !yield(dlg, nil) {
				return
			}
		}
	}
}

func (d *DSDelegationStore) Put(ctx context.Context, dlg ucan.Delegation) error {
	b, err := delegation.Encode(dlg)
	if err != nil {
		return err
	}
	if err := d.ds.Put(ctx, datastore.NewKey(dlg.Link().String()), b); err != nil {
		return err
	}
	return d.ds.Put(ctx, queryKey(dlg), b)
}

var _ Store = (*DSDelegationStore)(nil)

func queryKey(d ucan.Delegation) datastore.Key {
	aud := d.Audience().DID().String()
	cmd := sanitizeCommand(d.Command())
	var sub string
	if d.Subject() == nil {
		sub = NullSubject // powerline delegation
	} else {
		sub = d.Subject().DID().String()
	}
	key := fmt.Sprintf("%s/%s/%s/%s", aud, cmd, sub, d.Link().String())
	return datastore.NewKey(key)
}

func sanitizeCommand(cmd ucan.Command) string {
	return strings.ReplaceAll(string(cmd), "/", "~")
}
