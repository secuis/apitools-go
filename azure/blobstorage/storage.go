package blobstorage

import (
	"context"
	"io"
)

// BlobStorage provides access to multiple storage accounts at once
// Calling a function essentially forwards the request to the underlying account
// If no such account is found, ErrUnknownStorageAccount is returned
type Storage interface {
	BlobReader(ctx context.Context, account string, container string, blob string) (io.ReadCloser, error)
	BlobBytes(ctx context.Context, account string, container string, blob string) ([]byte, error)
	ListBlobs(ctx context.Context, account string, container string, prefix string) ([]string, error)
	ListBlobsByPattern(ctx context.Context, account string, container string, pattern string) ([]string, error)
	UploadBlob(ctx context.Context, account string, container string, reader io.Reader, blobName string) error
}

type blobStorage struct {
	accounts map[string]StorageAccount
}

// A collection of storage accounts
// Send in confs for all blob storage accounts you want to connect to
func New(confs []*AccountConfig) (Storage, error) {
	s := &blobStorage{
		accounts: map[string]StorageAccount{},
	}

	for _, c := range confs {
		if sa, err := NewAccount(c); err == nil {
			s.accounts[c.Name] = sa
		} else {
			return nil, err
		}
	}

	return s, nil
}

func (bs *blobStorage) BlobReader(ctx context.Context, account string, container string, blob string) (io.ReadCloser, error) {
	acc, ok := bs.accounts[account]

	if !ok {
		return nil, ErrUnknownStorageAccount
	}

	return acc.BlobReader(ctx, container, blob)
}

func (bs *blobStorage) BlobBytes(ctx context.Context, account string, container string, blob string) ([]byte, error) {
	acc, ok := bs.accounts[account]

	if !ok {
		return nil, ErrUnknownStorageAccount
	}

	return acc.BlobBytes(ctx, container, blob)
}

func (bs *blobStorage) ListBlobs(ctx context.Context, account string, container string, prefix string) ([]string, error) {
	acc, ok := bs.accounts[account]

	if !ok {
		return nil, ErrUnknownStorageAccount
	}

	return acc.ListBlobs(ctx, container, prefix)
}

func (bs *blobStorage) ListBlobsByPattern(ctx context.Context, account string, container string, pattern string) ([]string, error) {
	acc, ok := bs.accounts[account]

	if !ok {
		return nil, ErrUnknownStorageAccount
	}

	return acc.ListBlobsByPattern(ctx, container, pattern)
}

func (bs *blobStorage) UploadBlob(ctx context.Context, account string, container string, reader io.Reader, blobName string) error {
	acc, ok := bs.accounts[account]

	if !ok {
		return ErrUnknownStorageAccount
	}

	return acc.UploadBlob(ctx, container, reader, blobName)
}