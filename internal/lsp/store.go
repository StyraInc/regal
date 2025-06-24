package lsp

import (
	"context"
	"errors"
	"fmt"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/storage"
	"github.com/open-policy-agent/opa/v1/storage/inmem"

	"github.com/styrainc/roast/pkg/encoding"
)

func NewRegalStore() storage.Store {
	return inmem.NewFromObjectWithOpts(map[string]any{
		"workspace": map[string]any{
			"parsed": map[string]any{},
			// should map[string][]string{}, but since we don't round trip on write,
			// we'll need to conform to the most basic "JSON" format understood by the store
			"defined_refs": map[string]any{},
		},
	}, inmem.OptRoundTripOnWrite(false), inmem.OptReturnASTValuesOnRead(true))
}

func transact(ctx context.Context, store storage.Store, writeMode bool, op func(txn storage.Transaction) error) error {
	var params storage.TransactionParams
	if writeMode {
		params = storage.WriteParams
	}

	txn, err := store.NewTransaction(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	success := false
	defer func() {
		if !success {
			store.Abort(ctx, txn)
		}
	}()

	if err := op(txn); err != nil {
		return err
	}

	if err := store.Commit(ctx, txn); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	success = true

	return nil
}

func PutFileMod(ctx context.Context, store storage.Store, fileURI string, mod *ast.Module) error {
	return transact(ctx, store, true, func(txn storage.Transaction) error {
		var stErr *storage.Error

		var contents any

		contents = mod

		var modMap map[string]any

		// in extremely rare cases, this will fail:
		// https://github.com/StyraInc/regal/issues/1446
		// When it does, we let the store handle the conversion to a map.
		if err := encoding.JSONRoundTrip(mod, &modMap); err != nil {
			contents = modMap
		}

		err := store.Write(ctx, txn, storage.ReplaceOp, storage.Path{"workspace", "parsed", fileURI}, contents)
		if errors.As(err, &stErr) && stErr.Code == storage.NotFoundErr {
			if err = store.Write(ctx, txn, storage.AddOp, storage.Path{"workspace", "parsed", fileURI}, contents); err != nil {
				return fmt.Errorf("failed to init module in store: %w", err)
			}
		}

		if err != nil {
			return fmt.Errorf("failed to replace module in store: %w", err)
		}

		return nil
	})
}

func RemoveFileMod(ctx context.Context, store storage.Store, fileURI string) error {
	return transact(ctx, store, true, func(txn storage.Transaction) error {
		var stErr *storage.Error

		_, err := store.Read(ctx, txn, storage.Path{"workspace", "parsed", fileURI})
		if errors.As(err, &stErr) && stErr.Code == storage.NotFoundErr {
			// nothing to do
			return nil
		}

		if err != nil {
			return fmt.Errorf("failed to read module from store: %w", err)
		}

		if err = store.Write(ctx, txn, storage.RemoveOp, storage.Path{"workspace", "parsed", fileURI}, nil); err != nil {
			return fmt.Errorf("failed to remove module from store: %w", err)
		}

		return nil
	})
}

func PutFileRefs(ctx context.Context, store storage.Store, fileURI string, refs []string) error {
	return transact(ctx, store, true, func(txn storage.Transaction) error {
		var stErr *storage.Error

		err := store.Write(ctx, txn, storage.ReplaceOp, storage.Path{"workspace", "defined_refs", fileURI}, refs)
		if errors.As(err, &stErr) && stErr.Code == storage.NotFoundErr {
			err = store.Write(ctx, txn, storage.AddOp, storage.Path{"workspace", "defined_refs", fileURI}, refs)
			if err != nil {
				return fmt.Errorf("failed to init refs in store: %w", err)
			}
		}

		if err != nil {
			return fmt.Errorf("failed to replace refs in store: %w", err)
		}

		return nil
	})
}
