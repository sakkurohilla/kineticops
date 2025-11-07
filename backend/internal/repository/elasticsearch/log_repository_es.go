package elasticsearch

import "context"

// This file was previously an Elasticsearch repository implementation.
// The functionality has been intentionally removed. Keep a minimal stub so
// builds don't fail if any references remain temporarily. The user asked to
// remove Elasticsearch integration; we will remove references and dependencies
// next.

// EnsureIndexTemplate is a no-op stub (previously created ILM/index templates).
func EnsureIndexTemplate(ctx context.Context) error {
	return nil
}
