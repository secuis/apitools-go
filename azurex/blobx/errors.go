package blobx

import "github.com/pkg/errors"

var (
	ErrUnknownStorageAccount = errors.New("unknown storage account")
	ErrUploadFailed          = errors.New("blob upload failed")
	LockfileAlreadyExist     = errors.New("lockfile already exist")
)
