package blobx

import (
	"context"
	"io"
)

type BlobStorage struct {
	accounts map[string]*AccountConn
}

// A collection of storage accounts
// Send in confs for all blob storage accounts you want to connect to
func New(confs []*AccountConfig) (*BlobStorage, error) {
	s := &BlobStorage{
		accounts: map[string]*AccountConn{},
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

func (bs *BlobStorage) GetContainerSASURI(ctx context.Context, account string, container string, opts SASOptions) (string, error) {
	acc, ok := bs.accounts[account]

	if !ok {
		return "", ErrUnknownStorageAccount
	}

	return acc.GetContainerSASURI(ctx, container, opts)
}

func (bs *BlobStorage) GetBlobSASURI(ctx context.Context, account string, container string, blobName string, opts SASOptions) (string, error) {
	acc, ok := bs.accounts[account]

	if !ok {
		return "", ErrUnknownStorageAccount
	}

	return acc.GetBlobSASURI(ctx, container, blobName, opts)
}

func (bs *BlobStorage) BlobReader(ctx context.Context, account string, container string, blob string) (io.ReadCloser, error) {
	acc, ok := bs.accounts[account]

	if !ok {
		return nil, ErrUnknownStorageAccount
	}

	return acc.BlobReader(ctx, container, blob)
}

func (bs *BlobStorage) BlobBytes(ctx context.Context, account string, container string, blob string) ([]byte, error) {
	acc, ok := bs.accounts[account]

	if !ok {
		return nil, ErrUnknownStorageAccount
	}

	return acc.BlobBytes(ctx, container, blob)
}

func (bs *BlobStorage) ListBlobs(ctx context.Context, account string, container string, prefix string) ([]string, error) {
	acc, ok := bs.accounts[account]

	if !ok {
		return nil, ErrUnknownStorageAccount
	}

	return acc.ListBlobs(ctx, container, prefix)
}

func (bs *BlobStorage) ListBlobsByPattern(ctx context.Context, account string, container string, pattern string) ([]string, error) {
	acc, ok := bs.accounts[account]

	if !ok {
		return nil, ErrUnknownStorageAccount
	}

	return acc.ListBlobsByPattern(ctx, container, pattern)
}

// this method will handle acquire and release of the lease of the file
// if you already have the lease - then send in the leaseID
func (bs *BlobStorage) TruncateBlob(ctx context.Context, account string, container string, reader io.Reader, blobName string, leaseId string) error {
	acc, ok := bs.accounts[account]

	if !ok {
		return ErrUnknownStorageAccount
	}

	return acc.TruncateBlob(ctx, container, reader, blobName, leaseId)
}

// this method will handle acquire and release of the lease of the file
// if you already have the lease - then send in the leaseID
func (bs *BlobStorage) AppendBlob(ctx context.Context, account string, container string, reader io.Reader, blobName string, leaseId string) error {
	acc, ok := bs.accounts[account]

	if !ok {
		return ErrUnknownStorageAccount
	}

	return acc.AppendBlob(ctx, container, reader, blobName, leaseId)
}

// returns the leaseId for the file, and error if one occurred
func (bs *BlobStorage) AcquireLease(ctx context.Context, account string, container string, blobName string) (string, error) {
	acc, ok := bs.accounts[account]

	if !ok {
		return "", ErrUnknownStorageAccount
	}

	return acc.AcquireLease(ctx, container, blobName)
}

func (bs *BlobStorage) ReleaseLease(ctx context.Context, account string, container string, blobName string, leaseId string) error {
	acc, ok := bs.accounts[account]

	if !ok {
		return ErrUnknownStorageAccount
	}

	return acc.ReleaseLease(ctx, container, blobName, leaseId)
}