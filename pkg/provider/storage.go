package provider

import (
	"context"

	"github.com/0xsj/nexus-go/pkg/storage"
	"github.com/0xsj/nexus-go/pkg/storage/s3"
)

// NewStorage creates a new storage instance based on configuration
func ProvideStorage(ctx context.Context, cfg storage.Config) (storage.Storage, error) {
	return s3.New(ctx, cfg)
}
